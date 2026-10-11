#!/usr/bin/env bash
# Approve the CI runs that wait on the head of every open release pull request.
# The runs of a pull request opened with the repository token wait for an
# approval and report no check until then, and main requires the CI success
# check.
#   usage: tools/release-pr-ci.sh OWNER/REPO
#   needs: gh with a token that may read pull requests and write actions
#   env:   RELEASE_PR_CI_TRIES (default 6), RELEASE_PR_CI_WAIT seconds (default 10)
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: ${0##*/} OWNER/REPO" >&2
  exit 2
fi
REPO="$1"
GH="${GH_BIN:-gh}"
TRIES="${RELEASE_PR_CI_TRIES:-6}"
WAIT="${RELEASE_PR_CI_WAIT:-10}"
if [[ ! "$TRIES" =~ ^[1-9][0-9]?$ ]] || [[ ! "$WAIT" =~ ^[0-9]{1,3}$ ]]; then
  echo "error: unexpected RELEASE_PR_CI_TRIES or RELEASE_PR_CI_WAIT" >&2
  exit 2
fi

HEADS="$("$GH" pr list --repo "$REPO" --state open \
  --author 'app/github-actions' --label 'autorelease: pending' \
  --json headRefName,headRefOid --jq '.[] | "\(.headRefName) \(.headRefOid)"')"
if [ -z "$HEADS" ]; then
  echo "No open release pull request."
  exit 0
fi

# Prints "ID CONCLUSION BRANCH REPOSITORY" for every pull request run on a head.
runs_on() {
  "$GH" api "repos/$REPO/actions/runs?event=pull_request&head_sha=$1&per_page=100" \
    --jq '.workflow_runs[] | "\(.id) \(.conclusion // "none") \(.head_branch) \(.head_repository.full_name)"'
}

while read -r branch sha; do
  if [[ ! "$sha" =~ ^[0-9a-f]{40}$ ]]; then
    echo "error: unexpected head for $branch: $sha" >&2
    exit 1
  fi
  # The runs appear a moment after release-please has pushed the branch.
  runs=""
  for ((try = 1; try <= TRIES; try++)); do
    runs="$(runs_on "$sha")"
    if [ -n "$runs" ]; then
      break
    fi
    if [ "$try" -lt "$TRIES" ]; then
      sleep "$WAIT"
    fi
  done
  if [ -z "$runs" ]; then
    echo "error: no pull request run on $branch at $sha after $TRIES tries" >&2
    exit 1
  fi
  while read -r id conclusion run_branch run_repo; do
    if [[ ! "$id" =~ ^[0-9]+$ ]]; then
      echo "error: unexpected run on $sha: $id" >&2
      exit 1
    fi
    # Only a run of this repository's own release branch is approved.
    if [ "$conclusion" != action_required ] || [ "$run_branch" != "$branch" ] || [ "$run_repo" != "$REPO" ]; then
      continue
    fi
    "$GH" api --method POST "repos/$REPO/actions/runs/$id/approve" > /dev/null
    echo "Approved run $id on $branch at $sha."
  done <<< "$runs"
done <<< "$HEADS"
