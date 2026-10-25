#!/bin/sh
# Context budget report for the Claude Code knowledge base. POSIX sh and awk only.
# Measures every knowledge-base file, classes it by how often it is loaded, and
# compares it with the ceilings in .claude/context-budget.conf. It never edits anything.
#
# Usage: sh budget.sh [--root DIR] [--all] [--check]
#   --root DIR   project root (default: the git top level, else the current directory)
#   --all        list every measured file, not only the always-loaded ones and the overruns
#   --check      exit 1 if anything is over its ceiling or a pointer is broken

ROOT=""; ALL=0; CHECK=0
while [ $# -gt 0 ]; do
  case "$1" in
    --root) ROOT=$2; shift 2 ;;
    --all) ALL=1; shift ;;
    --check) CHECK=1; shift ;;
    -h|--help) sed -n '2,12p' "$0"; exit 0 ;;
    *) echo "unknown option: $1" >&2; exit 2 ;;
  esac
done

[ -n "$ROOT" ] || ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$ROOT" || { echo "cannot enter $ROOT" >&2; exit 2; }
CONF=.claude/context-budget.conf
[ -f "$CONF" ] || { echo "missing $CONF" >&2; exit 2; }

TMP=$(mktemp 2>/dev/null || echo "${TMPDIR:-/tmp}/budget.$$")
BROKEN="$TMP.broken"; OTHER="$TMP.other"
: > "$TMP"; : > "$BROKEN"; : > "$OTHER"
trap 'rm -f "$TMP" "$BROKEN" "$OTHER"' EXIT

# Ceiling for a key; 0 when absent.
lim() {
  v=$(grep "^$1=" "$CONF" | head -n 1 | cut -d= -f2 | tr -d ' \r')
  echo "${v:-0}"
}

# Estimated tokens for a file: bytes / 4.
tok() { echo $(( $(wc -c < "$1" | tr -d ' ') / 4 )); }

# Frontmatter facts: total lines, frontmatter end line, has "paths:", description characters.
fm_info() {
  awk '
    { sub(/\r$/, "") }
    NR == 1 && $0 == "---" { infm = 1; next }
    infm && $0 == "---" { infm = 0; fmend = NR; next }
    infm {
      if ($0 ~ /^paths:/) paths = 1
      if ($0 ~ /^description:/) {
        v = $0; sub(/^description:[ \t]*/, "", v)
        if (v ~ /^[>|][-+]?$/) v = ""
        desc = length(v); indesc = 1; next
      }
      if (indesc) {
        if ($0 ~ /^[ \t]/) { v = $0; sub(/^[ \t]+/, "", v); desc += length(v) + 1 }
        else indesc = 0
      }
    }
    END { printf "%d %d %d %d\n", NR, fmend + 0, paths + 0, desc + 0 }
  ' "$1"
}

# class|kind|unit|size|limit|tokens|path
row() { printf '%s|%s|%s|%s|%s|%s|%s\n' "$1" "$2" "$3" "$4" "$5" "$6" "$7" >> "$TMP"; }

for f in CLAUDE.md .claude/CLAUDE.md; do
  [ -f "$f" ] || continue
  set -- $(fm_info "$f")
  row always claude lines "$1" "$(lim claude_md_lines)" "$(tok "$f")" "$f"
done

for f in .claude/rules/*.md; do
  [ -f "$f" ] || continue
  set -- $(fm_info "$f")
  if [ "$3" = 1 ]; then
    row scoped rule lines "$1" "$(lim rule_scoped_lines)" "$(tok "$f")" "$f"
  else
    row always rule lines "$1" "$(lim rule_always_lines)" "$(tok "$f")" "$f"
  fi
done

for f in .claude/skills/*/SKILL.md; do
  [ -f "$f" ] || continue
  set -- $(fm_info "$f")
  row always skill-desc chars "$4" "$(lim skill_description_chars)" $(( $4 / 4 )) "$f"
  row invoke skill lines $(( $1 - $2 )) "$(lim skill_body_lines)" "$(tok "$f")" "$f"
done

for f in .claude/agents/*.md; do
  [ -f "$f" ] || continue
  set -- $(fm_info "$f")
  row always agent-desc chars "$4" "$(lim agent_description_chars)" $(( $4 / 4 )) "$f"
  row invoke agent lines $(( $1 - $2 )) "$(lim agent_body_lines)" "$(tok "$f")" "$f"
done

if [ -d .claude/skills ]; then
  find .claude/skills -path '*/references/*' -name '*.md' | sort | while read -r f; do
    set -- $(fm_info "$f")
    row demand reference lines "$1" "$(lim reference_lines)" "$(tok "$f")" "$f"
  done
fi

if [ -d .claude/docs ]; then
  find .claude/docs -name '*.md' | sort | while read -r f; do
    set -- $(fm_info "$f")
    row demand doc lines "$1" "$(lim doc_lines)" "$(tok "$f")" "$f"
  done
fi

find . \( -path ./.git -o -name node_modules \) -prune -o -path '*/deltas/*' \
  \( -name '*.spec.md' -o -name '*.design.md' \) -print | sed 's|^\./||' | sort | while read -r f; do
  set -- $(fm_info "$f")
  case "$f" in
    *.spec.md) row demand spec lines "$1" "$(lim spec_lines)" "$(tok "$f")" "$f" ;;
    *) row demand design lines "$1" "$(lim design_lines)" "$(tok "$f")" "$f" ;;
  esac
done

# Pointers from a skill into its own references/ must resolve.
for d in .claude/skills/*/; do
  [ -d "$d" ] || continue
  for f in "$d"SKILL.md "$d"references/*.md; do
    [ -f "$f" ] || continue
    grep -o 'references/[A-Za-z0-9._/-]*\.md' "$f" | sort -u | while read -r p; do
      [ -f "$d$p" ] && continue
      owner=""
      for s in .claude/skills/*/; do [ -f "$s$p" ] && owner=$s; done
      if [ -n "$owner" ]; then echo "$f -> $p (exists only in $owner)" >> "$OTHER"
      else echo "$f -> $p" >> "$BROKEN"; fi
    done
  done
done

SORTED="$TMP.sorted"
sort -t'|' -k1,1 -k6,6nr "$TMP" > "$SORTED"
rm -f "$TMP"; TMP="$SORTED"; trap 'rm -f "$TMP" "$BROKEN" "$OTHER"' EXIT

if [ ! -s "$TMP" ]; then echo "nothing to measure under $ROOT"; exit 0; fi

ALWAYS_LIMIT=$(lim always_total_tokens)
ALWAYS_TOTAL=$(awk -F'|' '$1 == "always" { s += $6 } END { print s + 0 }' "$TMP")
OVER=$(awk -F'|' '$5 > 0 && $4 > $5' "$TMP" | wc -l | tr -d ' ')
NBROKEN=$(wc -l < "$BROKEN" | tr -d ' ')
NOTHER=$(wc -l < "$OTHER" | tr -d ' ')

fmt='  %-11s %-6s %5d / %-5s ~%5dt  %s\n'
echo "ALWAYS LOADED (every turn)"
awk -F'|' -v f="$fmt" '$1 == "always" { printf f, $2, $3, $4, ($5 > 0 ? $5 : "-"), $6, $7 }' "$TMP"
status=ok
[ "$ALWAYS_LIMIT" -gt 0 ] && [ "$ALWAYS_TOTAL" -gt "$ALWAYS_LIMIT" ] && { status=OVER; OVER=$(( OVER + 1 )); }
echo "  total ~${ALWAYS_TOTAL}t of ${ALWAYS_LIMIT}t  [$status]"

if [ "$ALL" = 1 ]; then
  echo
  echo "EVERYTHING MEASURED"
  awk -F'|' -v f="  %-8s $fmt" '{ printf f, $1, $2, $3, $4, ($5 > 0 ? $5 : "-"), $6, $7 }' "$TMP" | sed 's/^   *//;s/^/  /'
fi

echo
echo "OVER A CEILING"
awk -F'|' -v f="$fmt" '$5 > 0 && $4 > $5 { printf f, $2, $3, $4, $5, $6, $7; n++ } END { if (!n) print "  none" }' "$TMP"

echo
if [ "$NBROKEN" -gt 0 ]; then
  echo "BROKEN POINTERS"
  sed 's/^/  /' "$BROKEN"
else
  echo "BROKEN POINTERS: none"
fi
if [ "$NOTHER" -gt 0 ]; then
  echo
  echo "AMBIGUOUS POINTERS (resolve only inside another skill)"
  sed 's/^/  /' "$OTHER"
fi

[ "$CHECK" = 1 ] && { [ "$OVER" -gt 0 ] || [ "$NBROKEN" -gt 0 ]; } && exit 1
exit 0
