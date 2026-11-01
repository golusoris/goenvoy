#!/usr/bin/env bash
# Verify that GitHub Actions workflow badges and links in top-level Markdown
# files point at workflow files that exist under .github/workflows.
#   usage: tools/readme-badges.sh [root]
set -euo pipefail

if [ $# -gt 1 ]; then
  echo "usage: ${0##*/} [root]" >&2
  exit 2
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT="${1:-$REPO_ROOT}"
ROOT="${ROOT%/}"

if [ ! -d "$ROOT" ]; then
  echo "error: directory $ROOT does not exist" >&2
  exit 1
fi

shopt -s nullglob
files=()
for f in "$ROOT"/*.md; do
  [ -f "$f" ] || continue
  files+=("$f")
done

# Pattern for badge image: actions/workflows/<file>/badge.svg
badge_re='actions/workflows/([^/?#[:space:]()\"'\''<>]+)/badge\.svg'
# Pattern for workflow link: https://github.com/<owner>/<repo>/actions/workflows/<file>
link_re='https?://github\.com/[^/]+/[^/]+/actions/workflows/([^/?#[:space:]()\"'\''<>]+)'

misses=0
ok_refs=0

for f in "${files[@]}"; do
  rel="${f#"$ROOT"/}"
  readarray -t lines <"$f"
  for ((lnum = 1; lnum <= ${#lines[@]}; lnum++)); do
    line="${lines[lnum - 1]}"
    [[ "$line" == *actions/workflows* ]] || continue

    cur="$line"
    for ((i = 0; i < 100; i++)); do
      if [[ "$cur" =~ $badge_re ]]; then
        matched="${BASH_REMATCH[0]}"
        wf="${BASH_REMATCH[1]}"
        if [ -f "$ROOT/.github/workflows/$wf" ]; then
          ok_refs=$((ok_refs + 1))
        else
          printf 'readme-badges: %s:%s: workflow %s not in .github/workflows\n' "$rel" "$lnum" "$wf" >&2
          misses=$((misses + 1))
        fi
        cur="${cur#*"$matched"}"
      else
        break
      fi
    done

    cur="$line"
    for ((i = 0; i < 100; i++)); do
      if [[ "$cur" =~ $link_re ]]; then
        matched="${BASH_REMATCH[0]}"
        wf="${BASH_REMATCH[1]}"
        after="${cur#*"$matched"}"
        # Why: a slash indicates subpaths such as /badge.svg or /runs rather than the workflow file.
        if [[ "$after" != /* ]]; then
          if [ -f "$ROOT/.github/workflows/$wf" ]; then
            ok_refs=$((ok_refs + 1))
          else
            printf 'readme-badges: %s:%s: workflow %s not in .github/workflows\n' "$rel" "$lnum" "$wf" >&2
            misses=$((misses + 1))
          fi
        fi
        cur="$after"
      else
        break
      fi
    done
  done
done

if [ "$misses" -gt 0 ]; then
  exit 1
fi

echo "readme-badges: $ok_refs workflow references ok"
