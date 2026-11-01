#!/bin/sh
# Verification coverage report for specs (POSIX sh and awk, read-only, no model).
# Reads the verification tag on every invariant and acceptance criterion of each
# <slice>.spec.md, counts them by kind, and checks that every test link resolves.
#
# Usage: sh links.sh [--root DIR] [--check] [spec ...]
#   --root DIR  project root (default: the git top level, else the current directory)
#   --check     exit 1 if any test link is broken
#   spec ...    only these spec files (default: every <slice>.spec.md under a deltas/ folder)

ROOT=""; CHECK=0; SPECS=""
while [ $# -gt 0 ]; do
  case $1 in
    --root) ROOT=$2; shift ;;
    --check) CHECK=1 ;;
    -*) sed -n '2,/^$/{/^$/!p;}' "$0"; exit 2 ;;
    *) SPECS="$SPECS $1" ;;
  esac
  shift
done

cd "${ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}" || exit 2
[ -n "$SPECS" ] || SPECS=$(find . \( -path ./.git -o -name node_modules \) -prune -o \
  -path '*/deltas/*' -name '*.spec.md' -print | sed 's|^\./||' | sort)
[ -n "$SPECS" ] || { echo "no specs found"; exit 0; }

TAB=$(printf '\t')

# One line per claim in Invariants (I) and Acceptance criteria (A): id, kind, argument.
claims() {
  awk '
    function flush(   tag, i, k, a) {
      if (!item) return
      n++
      if (match(text, /\[(test|check|manual|none)(:[^]]*)?\]/)) {
        tag = substr(text, RSTART + 1, RLENGTH - 2)
        i = index(tag, ":")
        k = i ? substr(tag, 1, i - 1) : tag
        a = i ? substr(tag, i + 1) : ""
        sub(/^[ \t]+/, "", a)
        printf "%s%d\t%s\t%s\n", sect, n, k, a
      } else printf "%s%d\tuntagged\t\n", sect, n
      item = 0; text = ""
    }
    { sub(/\r$/, "") }
    /^## / {
      flush(); n = 0
      sect = ($0 ~ /^## Invariants[ \t]*$/) ? "I" : ($0 ~ /^## Acceptance criteria[ \t]*$/) ? "A" : ""
      next
    }
    sect == "" { next }
    /^([0-9]+\.|-|\*)[ \t]/ { flush(); item = 1; text = $0; next }
    /^[ \t]*$/ { flush(); next }
    item { text = text " " $0 }
    END { flush() }
  ' "$1"
}

tally() {
  awk -F'\t' 'NF { c[$2]++; t++ }
    END { printf "%d claims | test %d | check %d | manual %d | none %d | untagged %d\n",
          t, c["test"], c["check"], c["manual"], c["none"], c["untagged"] }'
}

# A test tag is: path, optionally followed by the test name in double quotes.
# The path resolves from the project root or, in a multi-service repo, from the service root.
resolve() { # SPEC ID ARG
  path=${3%% *}
  name=""; case $3 in *\"*\"*) name=${3#*\"}; name=${name%\"*} ;; esac
  [ -n "$path" ] || { echo "  BROKEN $2: the test tag names no file"; return; }
  service=$(dirname "$(dirname "$1")")
  file=""; for c in "$path" "$service/$path"; do [ -f "$c" ] && { file=$c; break; }; done
  if [ -z "$file" ]; then echo "  BROKEN $2: test file not found: $path"
  elif [ -n "$name" ] && ! grep -qF -- "$name" "$file"; then
    echo "  BROKEN $2: test name not found in $path: \"$name\""
  fi
}

report() { # SPEC
  rows=$(claims "$1")
  echo "$1"
  echo "  $(echo "$rows" | tally)"
  echo "$rows" | while IFS=$TAB read -r id kind arg; do
    case $kind in
      test) resolve "$1" "$id" "$arg" ;;
      none) echo "  gap $id: no verification yet" ;;
      untagged) echo "  gap $id: no verification tag" ;;
    esac
  done
}

OUT=$(for spec in $SPECS; do report "$spec"; done)
echo "$OUT"
echo
echo "TOTAL $(echo "$SPECS" | wc -w | tr -d ' ') specs: $(for spec in $SPECS; do claims "$spec"; done | tally)"

BROKEN=$(( $(echo "$OUT" | grep -c '^  BROKEN') ))
[ "$CHECK" = 1 ] && [ "$BROKEN" -gt 0 ] && exit 1
exit 0
