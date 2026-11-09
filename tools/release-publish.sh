#!/usr/bin/env bash
# Publish a GitHub Release with its signed, attested assets.
#   usage:  tools/release-publish.sh stage OWNER/REPO TAG TITLE NOTES_FILE ASSET...
#           tools/release-publish.sh confirm OWNER/REPO TAG
#   modes:  stage:   verifies or creates a draft release, uploads and verifies all assets
#           confirm: verifies the release is published (draft is false)
#   needs:  gh, jq, timeout
#   env:    GH_BIN (default: gh)
#           RELEASE_PUBLISH_TIMEOUT (default: 300)
#   The workflow step runs stage, then gh release edit TAG --draft=false, then confirm.
set -euo pipefail

# usage: prints the usage line and exits 2.
usage() {
  echo "usage: ${0##*/} stage OWNER/REPO TAG TITLE NOTES_FILE ASSET..." >&2
  echo "       ${0##*/} confirm OWNER/REPO TAG" >&2
  exit 2
}

[ $# -ge 1 ] || usage
MODE="$1"
shift

if [ "$MODE" != "stage" ] && [ "$MODE" != "confirm" ]; then
  echo "error: unknown mode '$MODE', expected stage or confirm" >&2
  exit 2
fi

if [ "$MODE" = "stage" ]; then
  [ $# -ge 5 ] || usage
  REPO="$1"
  TAG="$2"
  TITLE="$3"
  NOTES="$4"
  shift 4
  ASSETS=("$@")
else
  [ $# -eq 2 ] || usage
  REPO="$1"
  TAG="$2"
fi

[ -n "$REPO" ] || usage
if [[ ! "$REPO" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "error: invalid repository '$REPO', expected OWNER/REPO" >&2
  exit 2
fi

[ -n "$TAG" ] || usage

if [ "$MODE" = "stage" ]; then
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
fi

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

if [ "$MODE" = "stage" ]; then
  ERR_TMP="$(mktemp "${TMPDIR:-/tmp}/release-publish.XXXXXX")"
  trap 'rm -f "$ERR_TMP"' EXIT

  is_draft=""
  view_rc=0
  is_draft="$(timeout "$TIMEOUT" "$GH" release view "$TAG" --repo "$REPO" --json isDraft --jq .isDraft 2>"$ERR_TMP")" || view_rc=$?

  if [ "$view_rc" -ne 0 ]; then
    if grep -q "release not found" "$ERR_TMP"; then
      gh_call "create release $TAG" release create "$TAG" --repo "$REPO" --draft --verify-tag --title "$TITLE" --notes-file "$NOTES"
      is_draft="true"
    else
      err_line="$(head -n 1 "$ERR_TMP")"
      echo "error: failed to view release $TAG: $err_line" >&2
      exit 1
    fi
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
    echo "release $TAG staged as a draft, assets verified" >&2
  else
    echo "release $TAG already published, assets refreshed" >&2
  fi

elif [ "$MODE" = "confirm" ]; then
  is_draft="$(gh_call "view release $TAG" release view "$TAG" --repo "$REPO" --json isDraft --jq .isDraft)"
  if [ "$is_draft" != "false" ]; then
    echo "error: release $TAG is still a draft" >&2
    exit 1
  fi
  echo "release $TAG published" >&2
fi
