#!/usr/bin/env bash
# Extract the release notes section for a given version from a changelog.
#   usage: tools/release-notes.sh CHANGELOG VERSION
#   needs: none
set -euo pipefail

usage() {
  echo "usage: ${0##*/} CHANGELOG VERSION" >&2
  exit 2
}

[ $# -eq 2 ] || usage

CHANGELOG="$1"
VERSION="$2"

if [ ! -f "$CHANGELOG" ]; then
  echo "error: $CHANGELOG: no such file" >&2
  exit 2
fi

prefix1="## [${VERSION}]"
exact2="## ${VERSION}"
prefix3="## ${VERSION} "

found=0
lines=()

while IFS= read -r line || [ -n "$line" ]; do
  if [ "$found" -eq 0 ]; then
    case "$line" in
      "$prefix1"* | "$exact2" | "$prefix3"*)
        found=1
        lines+=("$line")
        ;;
    esac
  else
    case "$line" in
      "## "*)
        break
        ;;
      *)
        lines+=("$line")
        ;;
    esac
  fi
done < "$CHANGELOG"

if [ "$found" -eq 0 ]; then
  echo "error: $CHANGELOG: no section for $VERSION" >&2
  exit 1
fi

while [ ${#lines[@]} -gt 0 ]; do
  last_idx=$((${#lines[@]} - 1))
  last_line="${lines[last_idx]}"
  if [ -z "${last_line//[[:space:]]/}" ]; then
    unset 'lines[last_idx]'
  else
    break
  fi
done

if [ ${#lines[@]} -gt 0 ]; then
  printf '%s\n' "${lines[@]}"
fi
