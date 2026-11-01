#!/usr/bin/env bash
# Verify that GitHub Actions workflow badges and links in top-level Markdown
# files point at workflow files that exist under .github/workflows.
#   usage: tools/readme-badges.sh [--repo owner/name] [root]
set -euo pipefail

REPO_SLUG=""
TARGET_ROOT=""

while [ $# -gt 0 ]; do
  case "$1" in
    --repo)
      if [ $# -lt 2 ]; then
        echo "usage: ${0##*/} [--repo owner/name] [root]" >&2
        exit 2
      fi
      REPO_SLUG="$2"
      shift 2
      ;;
    --repo=*)
      REPO_SLUG="${1#--repo=}"
      shift
      ;;
    -*)
      echo "usage: ${0##*/} [--repo owner/name] [root]" >&2
      exit 2
      ;;
    *)
      if [ -n "$TARGET_ROOT" ]; then
        echo "usage: ${0##*/} [--repo owner/name] [root]" >&2
        exit 2
      fi
      TARGET_ROOT="$1"
      shift
      ;;
  esac
done

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT="${TARGET_ROOT:-$REPO_ROOT}"
ROOT="${ROOT%/}"

if [ ! -d "$ROOT" ]; then
  echo "error: directory $ROOT does not exist" >&2
  exit 1
fi

if [ -z "$REPO_SLUG" ]; then
  origin_url=""
  if git -C "$REPO_ROOT" remote get-url origin >/dev/null 2>&1; then
    origin_url="$(git -C "$REPO_ROOT" remote get-url origin 2>/dev/null)"
  fi
  origin_url="${origin_url%.git}"
  origin_url="${origin_url%/}"
  if [[ "$origin_url" =~ [:/]([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)$ ]]; then
    REPO_SLUG="${BASH_REMATCH[1]}"
  fi
fi

if [ -z "$REPO_SLUG" ]; then
  echo "error: could not determine repository slug (use --repo owner/name)" >&2
  exit 2
fi

shopt -s nullglob
files=()
for f in "$ROOT"/*.md "$ROOT"/*.markdown; do
  if [ ! -f "$f" ]; then
    continue
  fi
  files+=("$f")
done

# Pattern for workflow references:
# 1) github.com/<owner>/<repo>/actions/workflows/<file>
# 2) actions/workflows/<file> (not preceded by github.com/<owner>/<repo>/)
# where <file> is [A-Za-z0-9._-]+
workflow_re="(github\.com/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)/)?actions/workflows/([A-Za-z0-9._-]+)"

misses=0
ok_refs=0

for f in "${files[@]}"; do
  rel="${f#"$ROOT"/}"
  readarray -t lines <"$f"
  for ((lnum = 1; lnum <= ${#lines[@]}; lnum++)); do
    line="${lines[lnum - 1]}"
    if [[ "$line" != *actions/workflows* ]]; then
      continue
    fi

    cur="$line"
    limit=$((${#line} + 1))
    for ((i = 0; i < limit; i++)); do
      if [[ "$cur" =~ $workflow_re ]]; then
        matched="${BASH_REMATCH[0]}"
        owner_repo="${BASH_REMATCH[2]}"
        wf="${BASH_REMATCH[3]}"
        cur="${cur#*"$matched"}"

        # Why: trailing dots are sentence punctuation; strip them without a loop.
        wf="${wf%"${wf##*[!.]}"}"

        # Why: ignore references to workflows hosted in other repositories.
        if [ -n "$owner_repo" ] && [ "${owner_repo,,}" != "${REPO_SLUG,,}" ]; then
          continue
        fi

        if [ -n "$wf" ] && [ -f "$ROOT/.github/workflows/$wf" ]; then
          ok_refs=$((ok_refs + 1))
        else
          printf 'readme-badges: %s:%s: workflow %s not in .github/workflows\n' "$rel" "$lnum" "$wf" >&2
          misses=$((misses + 1))
        fi
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
