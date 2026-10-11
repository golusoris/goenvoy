#!/usr/bin/env bash
# Guard between the release manifest and the Go module paths, and the tag list
# of a release run.
#   usage: tools/release-manifest.sh check [ROOT]
#          tools/release-manifest.sh tags PATHS_JSON [ROOT]
#   check: every manifest entry is a module whose path carries the major of
#          its version (v2 and up end in /vN, v0 and v1 carry no suffix).
#   tags:  prints the JSON array of <path>/v<version> for the released paths
#          release-please reports, after the same check on each of them.
#   needs: jq
set -euo pipefail

usage() {
  echo "usage: ${0##*/} check [ROOT] | tags PATHS_JSON [ROOT]" >&2
  exit 2
}

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-}"
PATHS='[]'
case "$MODE" in
  check)
    [ $# -le 2 ] || usage
    ROOT="${2:-$REPO_ROOT}"
    ;;
  tags)
    { [ $# -ge 2 ] && [ $# -le 3 ]; } || usage
    PATHS="$2"
    ROOT="${3:-$REPO_ROOT}"
    ;;
  *) usage ;;
esac

MANIFEST="$ROOT/.release-please-manifest.json"
if [ ! -f "$MANIFEST" ]; then
  echo "error: no release manifest at $MANIFEST" >&2
  exit 1
fi

# check_path PATH: prints the manifest version of PATH, or names why the
# module cannot be released at that version.
check_path() {
  local path="$1" version major module
  if ! version="$(jq -er --arg p "$path" '.[$p] | strings' "$MANIFEST")"; then
    echo "error: $path: not in the release manifest" >&2
    return 1
  fi
  if [[ ! "$version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$ ]]; then
    echo "error: $path: version $version is not X.Y.Z" >&2
    return 1
  fi
  major="${BASH_REMATCH[1]}"
  if [ ! -f "$ROOT/$path/go.mod" ]; then
    echo "error: $path: no go.mod" >&2
    return 1
  fi
  module="$(awk '$1 == "module" { print $2; exit }' "$ROOT/$path/go.mod")"
  # Go resolves a v2+ tag only for a module path that ends in the same /vN.
  if [ "$major" -ge 2 ]; then
    if [[ "$module" != */v"$major" ]]; then
      echo "error: $path: version $version needs a module path ending in /v$major, got $module" >&2
      return 1
    fi
  elif [[ "$module" =~ /v[0-9]+$ ]]; then
    echo "error: $path: version $version does not fit the module path $module" >&2
    return 1
  fi
  printf '%s\n' "$version"
}

if [ "$MODE" = check ]; then
  count=0
  failed=0
  while IFS= read -r path; do
    count=$((count + 1))
    check_path "$path" > /dev/null || failed=$((failed + 1))
  done < <(jq -r 'keys[]' "$MANIFEST")
  if [ "$count" -eq 0 ]; then
    echo "error: the release manifest lists no module" >&2
    exit 1
  fi
  if [ "$failed" -ne 0 ]; then
    echo "error: $failed of $count manifest entries do not match their module" >&2
    exit 1
  fi
  echo "release manifest: $count modules, every version matches its module path"
  exit 0
fi

if ! jq -e 'type == "array" and all(.[]; type == "string")' <<< "$PATHS" > /dev/null 2>&1; then
  echo "error: released paths are not a JSON array of strings: $PATHS" >&2
  exit 1
fi
TAGS=()
while IFS= read -r path; do
  version="$(check_path "$path")"
  TAGS+=("$path/v$version")
done < <(jq -r '.[]' <<< "$PATHS")
if [ ${#TAGS[@]} -eq 0 ]; then
  echo '[]'
else
  printf '%s\n' "${TAGS[@]}" | jq -R . | jq -cs .
fi
