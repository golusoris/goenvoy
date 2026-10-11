#!/usr/bin/env bash
# Dispatch CI on the head of every open release pull request that has no
# dispatched CI run yet. A pull request opened with the repository token gets
# no CI run of its own, and main requires the CI success check.
#   usage: tools/release-pr-ci.sh OWNER/REPO
#   needs: gh with a token that may read pull requests and write actions
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: ${0##*/} OWNER/REPO" >&2
  exit 2
fi
REPO="$1"
GH="${GH_BIN:-gh}"

HEADS="$("$GH" pr list --repo "$REPO" --state open \
  --author 'app/github-actions' --label 'autorelease: pending' \
  --json headRefName,headRefOid --jq '.[] | "\(.headRefName) \(.headRefOid)"')"
if [ -z "$HEADS" ]; then
  echo "No open release pull request."
  exit 0
fi

while read -r branch sha; do
  if [[ ! "$sha" =~ ^[0-9a-f]{40}$ ]]; then
    echo "error: unexpected head for $branch: $sha" >&2
    exit 1
  fi
  runs="$("$GH" api "repos/$REPO/actions/workflows/ci.yml/runs?event=workflow_dispatch&head_sha=$sha" --jq .total_count)"
  if [[ ! "$runs" =~ ^[0-9]+$ ]]; then
    echo "error: unexpected run count for $sha: $runs" >&2
    exit 1
  fi
  if [ "$runs" -gt 0 ]; then
    echo "$branch at $sha already has a dispatched CI run."
    continue
  fi
  "$GH" workflow run ci.yml --repo "$REPO" --ref "$branch"
  echo "Dispatched CI on $branch at $sha."
done <<< "$HEADS"
