#!/usr/bin/env bash
# Publish a GitHub Release with its signed, attested assets.
#   usage:  tools/release-publish.sh OWNER/REPO TAG TITLE NOTES_FILE ASSET...
#   needs:  gh, jq, timeout
#   env:    GH_BIN (default: gh)
#           RELEASE_PUBLISH_TIMEOUT (default: 300)
set -euo pipefail

# usage: prints the usage line and exits 2.
usage() {
  echo "usage: ${0##*/} OWNER/REPO TAG TITLE NOTES_FILE ASSET..." >&2
  exit 2
}

[ $# -ge 5 ] || usage

REPO="$1"
TAG="$2"
TITLE="$3"
NOTES="$4"
shift 4
ASSETS=("$@")

if [[ ! "$REPO" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "error: invalid repository '$REPO', expected OWNER/REPO" >&2
  exit 2
fi

[ -n "$TAG" ] || usage
[ -n "$TITLE" ] || usage

if [ ! -f "$NOTES" ]; then
  echo "error: notes file '$NOTES' does not exist" >&2
  exit 2
fi

for asset in "${ASSETS[@]}"; do
  if [ ! -f "$asset" ]; then
    echo "error: asset file '$asset' does not exist" >&2
    exit 2
  fi
done

TIMEOUT="${RELEASE_PUBLISH_TIMEOUT:-300}"
if [[ ! "$TIMEOUT" =~ ^[1-9][0-9]{0,3}$ ]]; then
  echo "error: invalid RELEASE_PUBLISH_TIMEOUT: '$TIMEOUT'" >&2
  exit 2
fi

GH="${GH_BIN:-gh}"

for cmd in "$GH" jq timeout; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: $cmd is required" >&2
    exit 1
  fi
done

# gh_call STEP ARGS...: runs a gh command under the deadline; on failure names STEP and exits 1.
gh_call() {
  local step="$1"
  shift
  timeout "$TIMEOUT" "$GH" "$@" || {
    echo "error: failed to $step" >&2
    exit 1
  }
}

is_draft=""
if ! is_draft="$(gh_call "view release $TAG" release view "$TAG" --repo "$REPO" --json isDraft --jq .isDraft 2>/dev/null)"; then
  gh_call "create release $TAG" release create "$TAG" --repo "$REPO" --draft --verify-tag --title "$TITLE" --notes-file "$NOTES"
  is_draft="true"
fi

if [ "$is_draft" != "true" ] && [ "$is_draft" != "false" ]; then
  echo "error: release $TAG unexpected draft state '$is_draft'" >&2
  exit 1
fi

gh_call "upload assets for $TAG" release upload "$TAG" --repo "$REPO" --clobber "${ASSETS[@]}"

assets_json="$(gh_call "verify assets for $TAG" release view "$TAG" --repo "$REPO" --json assets --jq '[.assets[].name]')"
for asset in "${ASSETS[@]}"; do
  asset_name="${asset##*/}"
  if ! jq -e --arg name "$asset_name" 'index($name) != null' <<< "$assets_json" >/dev/null 2>&1; then
    echo "error: release $TAG lacks asset $asset_name" >&2
    exit 1
  fi
done

if [ "$is_draft" = "true" ]; then
  gh_call "publish release $TAG" release edit "$TAG" --repo "$REPO" --draft=false
  recheck_draft="$(gh_call "recheck draft status for $TAG" release view "$TAG" --repo "$REPO" --json isDraft --jq .isDraft)"
  if [ "$recheck_draft" != "false" ]; then
    echo "error: release $TAG is still a draft after publish" >&2
    exit 1
  fi
else
  echo "release $TAG already published, assets refreshed" >&2
fi
