#!/usr/bin/env bash
# Tag the merged release pull requests and relabel them for publication.
#   usage:  tools/release-tag.sh OWNER/REPO
#   needs:  gh, jq, timeout
#   env:    GH_BIN (default: gh)
#           RELEASE_TAG_ROOT (default: repository root)
#           RELEASE_TAG_DRY_RUN (default: 0)
#           RELEASE_TAG_TIMEOUT (default: 60)
#   stdout: JSON array of tags in C-locale order
set -euo pipefail

# usage: prints the usage line and exits 2.
usage() {
  echo "usage: ${0##*/} OWNER/REPO" >&2
  exit 2
}

[ $# -eq 1 ] || usage
REPO="$1"
[ -n "$REPO" ] || usage
if [[ ! "$REPO" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "error: invalid repository '$REPO', expected OWNER/REPO" >&2
  exit 2
fi

TIMEOUT="${RELEASE_TAG_TIMEOUT:-60}"
if [[ ! "$TIMEOUT" =~ ^[1-9][0-9]{0,2}$ ]]; then
  echo "error: invalid RELEASE_TAG_TIMEOUT: '$TIMEOUT'" >&2
  exit 2
fi

RELEASE_BRANCH="release-please--branches--main"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GH="${GH_BIN:-gh}"
ROOT="${RELEASE_TAG_ROOT:-$(cd "$HERE/.." && pwd)}"
DRY_RUN="${RELEASE_TAG_DRY_RUN:-0}"

for cmd in "$GH" jq timeout; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: $cmd is required" >&2
    exit 1
  fi
done

# gh_read STEP ARGS...: runs a read-only gh call under the deadline; on failure names STEP and exits 1.
gh_read() {
  local step="$1"
  shift
  timeout "$TIMEOUT" "$GH" "$@" || {
    echo "error: failed to $step" >&2
    exit 1
  }
}

# gh_write METHOD ENDPOINT ARGS...: runs one write under the deadline, or prints it when RELEASE_TAG_DRY_RUN=1.
gh_write() {
  local method="$1" endpoint="$2"
  shift 2
  if [ "$DRY_RUN" = "1" ]; then
    local -a fields=()
    while [ $# -gt 0 ]; do
      if [ "$1" = "-f" ]; then
        shift
        fields+=("$1")
      else
        fields+=("$1")
      fi
      shift
    done
    if [ ${#fields[@]} -gt 0 ]; then
      printf 'dry-run: %s %s %s\n' "$method" "$endpoint" "${fields[*]}" >&2
    else
      printf 'dry-run: %s %s\n' "$method" "$endpoint" >&2
    fi
    return 0
  fi
  timeout "$TIMEOUT" "$GH" api --method "$method" "$endpoint" "$@" >/dev/null || {
    echo "error: failed to $method $endpoint" >&2
    exit 1
  }
}

# Why: gh pr list --label reads the search index, which lags behind a merge; the REST list with a head filter does not.
OWNER="${REPO%%/*}"
pr_lines="$(gh_read "list release pull requests for $REPO" api "repos/$REPO/pulls?state=closed&head=$OWNER:$RELEASE_BRANCH&per_page=20" --jq '.[] | select(.merged_at != null) | select(.head.ref == "'"$RELEASE_BRANCH"'") | select(any(.labels[]; .name == "autorelease: pending")) | "\(.number) \(.merge_commit_sha)"')"

prs=()
if [ -n "$pr_lines" ]; then
  while IFS= read -r pr_line || [ -n "$pr_line" ]; do
    [ -n "$pr_line" ] || continue
    if [[ ! "$pr_line" =~ ^([1-9][0-9]*)[[:space:]]+([0-9a-f]{40})$ ]]; then
      echo "error: invalid pull request entry: '$pr_line'" >&2
      exit 1
    fi
    prs+=("$pr_line")
  done <<< "$pr_lines"
fi

all_tags=()

for pr_entry in "${prs[@]}"; do
  [[ "$pr_entry" =~ ^([1-9][0-9]*)[[:space:]]+([0-9a-f]{40})$ ]]
  pr_num="${BASH_REMATCH[1]}"
  sha="${BASH_REMATCH[2]}"

  parents="$(gh_read "fetch commit $sha for PR $pr_num" api "repos/$REPO/commits/$sha" --jq '.parents | map(.sha) | join(" ")')"
  if [[ ! "$parents" =~ ^[0-9a-f]{40}$ ]]; then
    echo "error: release pull request $pr_num merge commit $sha must have exactly one parent, got: '$parents'" >&2
    exit 1
  fi
  parent="$parents"

  parent_manifest="$(gh_read "fetch parent manifest for PR $pr_num at $parent" api -H 'Accept: application/vnd.github.raw' "repos/$REPO/contents/.release-please-manifest.json?ref=$parent")"
  sha_manifest="$(gh_read "fetch manifest for PR $pr_num at $sha" api -H 'Accept: application/vnd.github.raw' "repos/$REPO/contents/.release-please-manifest.json?ref=$sha")"

  if ! jq -e 'type == "object" and all(.[]; type == "string")' <<< "$parent_manifest" >/dev/null 2>&1; then
    echo "error: PR $pr_num parent manifest at $parent is not a JSON object of strings" >&2
    exit 1
  fi
  if ! jq -e 'type == "object" and all(.[]; type == "string")' <<< "$sha_manifest" >/dev/null 2>&1; then
    echo "error: PR $pr_num manifest at $sha is not a JSON object of strings" >&2
    exit 1
  fi

  paths_json="$(jq -nc --argjson parent "$parent_manifest" --argjson sha "$sha_manifest" \
    '[$sha | to_entries[] | select(.value != $parent[.key]) | .key] | sort')"
  if [ "$(jq 'length' <<< "$paths_json")" -eq 0 ]; then
    echo "error: release pull request $pr_num changes no manifest version" >&2
    exit 1
  fi

  tags_json="$(bash "$HERE/release-manifest.sh" tags "$paths_json" "$ROOT")" || {
    echo "error: release-manifest.sh failed for PR $pr_num" >&2
    exit 1
  }

  mapfile -t paths < <(jq -r '.[]' <<< "$paths_json")
  mapfile -t tags < <(jq -r '.[]' <<< "$tags_json")

  if [ ${#paths[@]} -ne ${#tags[@]} ]; then
    echo "error: PR $pr_num: path count ${#paths[@]} does not match tag count ${#tags[@]}" >&2
    exit 1
  fi

  for i in "${!paths[@]}"; do
    p="${paths[$i]}"
    t="${tags[$i]}"
    sha_ver="$(jq -r --arg p "$p" '.[$p]' <<< "$sha_manifest")"
    want_tag="$p/v$sha_ver"
    if [ "$t" != "$want_tag" ]; then
      echo "error: tag $t does not match expected $want_tag in manifest at $sha (the checkout manifest moved)" >&2
      exit 1
    fi
  done

  for tag in "${tags[@]}"; do
    matching="$(gh_read "check matching refs for tag $tag" api "repos/$REPO/git/matching-refs/tags/$tag" --jq ".[] | select(.ref == \"refs/tags/$tag\") | \"\(.object.type) \(.object.sha)\"")"
    if [ -z "$matching" ]; then
      gh_write POST "repos/$REPO/git/refs" -f "ref=refs/tags/$tag" -f "sha=$sha"
    elif [ "$matching" = "commit $sha" ]; then
      echo "tag $tag already exists at commit $sha" >&2
    else
      echo "error: tag $tag exists at $matching, want commit $sha" >&2
      exit 1
    fi
    all_tags+=("$tag")
  done

  gh_write POST "repos/$REPO/issues/$pr_num/labels" -f "labels[]=autorelease: tagged"
  gh_write DELETE "repos/$REPO/issues/$pr_num/labels/autorelease:%20pending"
done

if [ ${#all_tags[@]} -eq 0 ]; then
  echo '[]'
else
  printf '%s\n' "${all_tags[@]}" | LC_ALL=C sort -u | jq -R . | jq -sc .
fi
