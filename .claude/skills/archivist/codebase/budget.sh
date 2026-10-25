#!/bin/sh
# Context budget report for the Claude Code knowledge base (POSIX sh and awk, read-only).
# Classes every knowledge-base file by how often it is loaded and compares it with the
# ceilings in .claude/budget.conf.
#
# Usage: sh budget.sh [--root DIR] [--all] [--check]
#   --root DIR  project root (default: the git top level, else the current directory)
#   --all       also list every measured file
#   --check     exit 1 if anything is over a ceiling or a pointer is broken

ROOT=""; ALL=0; CHECK=0
while [ $# -gt 0 ]; do
  case $1 in
    --root) ROOT=$2; shift ;;
    --all) ALL=1 ;;
    --check) CHECK=1 ;;
    *) sed -n '2,/^$/{/^$/!p;}' "$0"; exit 2 ;;
  esac
  shift
done

cd "${ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}" || exit 2
[ -f .claude/budget.conf ] || { echo "missing .claude/budget.conf" >&2; exit 2; }
eval "$(tr -d '\r' < .claude/budget.conf)"

limit() { eval "echo \${$1:-0}"; }
est() { echo $(( $1 * 10 / 33 )); }   # about 3.3 characters per token, calibrated against /context
tokens() { est "$(wc -c < "$1")"; }

# Prints: lines after the frontmatter, 1 if it has "paths:", characters of its description.
frontmatter() {
  awk '
    { sub(/\r$/, "") }
    NR == 1 && $0 == "---" { infm = 1; next }
    infm && $0 == "---" { infm = 0; fmend = NR; next }
    infm && /^paths:/ { paths = 1 }
    infm && /^description:/ {
      v = $0; sub(/^description:[ \t]*/, "", v)
      if (v ~ /^[>|][-+]?$/) v = ""
      desc = length(v); indesc = 1; next
    }
    infm && indesc {
      if ($0 ~ /^[ \t]/) { v = $0; sub(/^[ \t]+/, "", v); desc += length(v) + 1 } else indesc = 0
    }
    END { print NR - fmend, paths + 0, desc + 0 }
  ' "$1"
}

# emit CLASS KIND UNIT SIZE LIMIT-KEY FILE [TOKENS]  ->  class|kind|unit|size|limit|tokens|file
emit() { echo "$1|$2|$3|$4|$(limit "$5")|${7:-$(tokens "$6")}|$6"; }

# measure KIND FILE: emit the rows for one file.
measure() {
  set -- "$1" "$2" $(frontmatter "$2")   # $3 body lines, $4 has paths, $5 description chars
  case $1 in
    claude) emit always claude lines "$3" claude_lines "$2" ;;
    rule)
      if [ "$4" = 1 ]; then emit scoped rule lines "$3" rule_scoped_lines "$2"
      else emit always rule lines "$3" rule_always_lines "$2"; fi ;;
    skill|agent)
      emit always "$1-desc" chars "$5" "$1_desc_chars" "$2" "$(est "$5")"
      emit invoke "$1" lines "$3" "$1_body_lines" "$2" ;;
    *) emit demand "$1" lines "$3" "$1_lines" "$2" ;;
  esac
}

rows() {
  for f in CLAUDE.md .claude/CLAUDE.md; do [ -f "$f" ] && measure claude "$f"; done
  for f in .claude/rules/*.md; do [ -f "$f" ] && measure rule "$f"; done
  for f in .claude/skills/*/SKILL.md; do [ -f "$f" ] && measure skill "$f"; done
  for f in .claude/agents/*.md; do [ -f "$f" ] && measure agent "$f"; done
  for f in $(find .claude/skills -path '*/references/*' -name '*.md' 2>/dev/null); do measure reference "$f"; done
  for f in $(find .claude/docs -name '*.md' 2>/dev/null); do measure doc "$f"; done
  for f in $(find . \( -path ./.git -o -name node_modules \) -prune -o -path '*/deltas/*' \
      \( -name '*.spec.md' -o -name '*.design.md' \) -print); do
    case $f in *.spec.md) measure spec "${f#./}" ;; *) measure design "${f#./}" ;; esac
  done
}

# A skill's `references/...md` pointers must exist in its own folder. One that exists only in
# another skill resolves, but is ambiguous; a full path from the project root is unambiguous.
pointers() {
  for d in .claude/skills/*/; do
    for f in "$d"SKILL.md "$d"references/*.md; do
      [ -f "$f" ] || continue
      for p in $(grep -o '[A-Za-z0-9._/-]*references/[A-Za-z0-9._/-]*\.md' "$f" | sort -u); do
        case $p in
          references/*) [ -f "$d$p" ] && continue ;;
          *) [ -f "$p" ] && continue ;;
        esac
        owner=$(ls -d .claude/skills/*/"$p" 2>/dev/null | head -n 1)
        if [ -n "$owner" ]; then echo "ambiguous|$f -> $p (only in ${owner%/"$p"})"
        else echo "broken|$f -> $p"; fi
      done
    done
  done
}

ROWS=$(rows | sort -t'|' -k1,1 -k6,6nr)
[ -n "$ROWS" ] || { echo "nothing to measure"; exit 0; }
POINTERS=$(pointers)

table() { # all | always | over
  echo "$ROWS" | awk -F'|' -v show="$1" '
    show == "always" && $1 != "always" { next }
    show == "over" && !($5 > 0 && $4 > $5) { next }
    { printf "  %-8s %-11s %5d / %-5s ~%5dt  %s\n", $1, $2, $4, ($5 > 0 ? $5 : "-"), $6, $7; n++ }
    END { if (!n) print "  none" }'
}

TOTAL=$(echo "$ROWS" | awk -F'|' '$1 == "always" { s += $6 } END { print s + 0 }')
OVER=$(( $(echo "$ROWS" | awk -F'|' '$5 > 0 && $4 > $5' | wc -l) ))
BROKEN=$(( $(echo "$POINTERS" | grep -c '^broken|') ))
TOTAL_OVER=0; [ "${always_tokens:-0}" -gt 0 ] && [ "$TOTAL" -gt "$always_tokens" ] && TOTAL_OVER=1

echo "ALWAYS LOADED (every turn)"; table always
echo "  total ~${TOTAL}t of ${always_tokens:-0}t [$([ $TOTAL_OVER = 1 ] && echo OVER || echo ok)]"
echo; echo "OVER A CEILING"; table over
echo; echo "POINTERS"
[ -n "$POINTERS" ] || echo "  all resolve"
echo "$POINTERS" | sed -n 's/^broken|/  broken: /p; s/^ambiguous|/  ambiguous: /p'
[ "$ALL" = 1 ] && { echo; echo "EVERYTHING MEASURED"; table all; }

[ "$CHECK" = 1 ] && [ $(( OVER + TOTAL_OVER + BROKEN )) -gt 0 ] && exit 1
exit 0
