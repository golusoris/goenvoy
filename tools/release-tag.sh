#!/usr/bin/env bash
# Tag the merged release pull requests (tag) or relabel them after publication (relabel).
#   usage:  tools/release-tag.sh tag|relabel OWNER/REPO
#   needs:  gh, jq, timeout
#   env:    GH_BIN (default: gh)
#           RELEASE_TAG_DRY_RUN (default: 0)
#           RELEASE_TAG_TIMEOUT (default: 60)
#   stdout: mode tag:     JSON array of tags in C-locale order
#           mode relabel: compact JSON array of relabelled PR numbers in ascending numeric order
set -euo pipefail
# Why: plan_tags and tag_state run in command substitutions; without inherit_errexit a failed read inside them would print a state instead of stopping.
shopt -s inherit_errexit

# usage: prints the usage line and exits 2.
usage() {
  echo "usage: ${0##*/} tag|relabel OWNER/REPO" >&2
  exit 2
}

[ $# -eq 2 ] || usage
MODE="$1"
REPO="$2"

if [ "$MODE" != "tag" ] && [ "$MODE" != "relabel" ]; then
  echo "error: unknown mode '$MODE', expected tag or relabel" >&2
  exit 2
fi

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
DRY_RUN="${RELEASE_TAG_DRY_RUN:-0}"

for cmd in "$GH" jq timeout; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: $cmd is required" >&2
    exit 1
  fi
done

# Why: plan_tags checks tags against files read at the merge commit; this folder holds them and goes away on exit.
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

# gh_read STEP ARGS...: runs a read-only gh call under the deadline; on failure names STEP and exits 1.
gh_read() {
  local step="$1"
  shift
  timeout "$TIMEOUT" "$GH" "$@" || {
    echo "error: failed to $step" >&2
    exit 1
  }
}

# gh_try_write METHOD ENDPOINT ARGS...: runs one write under deadline, or prints warning and returns 1.
gh_try_write() {
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
    echo "warning: $method $endpoint failed" >&2
    return 1
  }
}

# gh_write METHOD ENDPOINT ARGS...: runs one write under deadline or exits 1 on failure.
gh_write() {
  local method="$1" endpoint="$2"
  gh_try_write "$@" || {
    echo "error: failed to $method $endpoint" >&2
    exit 1
  }
}

# merge_tree PR SHA MANIFEST PATHS_JSON: writes MANIFEST and the go.mod of each path in PATHS_JSON, read at SHA, into a new folder and prints its path.
merge_tree() {
  local pr_num="$1" sha="$2" manifest="$3" paths_json="$4"
  local tree p
  local -a paths
  tree="$(mktemp -d "$WORK_DIR/pr-$pr_num.XXXXXX")"
  printf '%s\n' "$manifest" > "$tree/.release-please-manifest.json"
  mapfile -t paths < <(jq -r '.[]' <<< "$paths_json")
  for p in "${paths[@]}"; do
    # Why: the path names a folder of the release commit and goes into the contents URL; the allow-list keeps it inside both.
    if [[ ! "$p" =~ ^[a-z0-9][a-z0-9_-]*(/[a-z0-9][a-z0-9_.-]*)*$ ]]; then
      echo "error: PR $pr_num: manifest path '$p' is not a module folder" >&2
      exit 1
    fi
    mkdir -p "$tree/$p"
    gh_read "fetch $p/go.mod for PR $pr_num at $sha" api -H 'Accept: application/vnd.github.raw' "repos/$REPO/contents/$p/go.mod?ref=$sha" > "$tree/$p/go.mod"
  done
  printf '%s\n' "$tree"
}

# plan_tags PR SHA: plans and prints tags for PR at SHA, one per line.
plan_tags() {
  local pr_num="$1" sha="$2"
  local parents parent parent_manifest sha_manifest paths_json tags_json tree
  local -a paths tags

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

  tree="$(merge_tree "$pr_num" "$sha" "$sha_manifest" "$paths_json")"
  tags_json="$(bash "$HERE/release-manifest.sh" tags "$paths_json" "$tree")" || {
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
    local p="${paths[$i]}"
    local t="${tags[$i]}"
    local sha_ver
    sha_ver="$(jq -r --arg p "$p" '.[$p]' <<< "$sha_manifest")"
    local want_tag="$p/v$sha_ver"
    if [ "$t" != "$want_tag" ]; then
      echo "error: tag $t does not match expected $want_tag in manifest at $sha" >&2
      exit 1
    fi
  done

  printf '%s\n' "${tags[@]}"
}

# tag_state TAG SHA: prints tag state (missing or ok) or exits 1 on conflict.
tag_state() {
  local tag="$1" sha="$2"
  local matching
  matching="$(gh_read "check matching refs for tag $tag" api "repos/$REPO/git/matching-refs/tags/$tag" --jq ".[] | select(.ref == \"refs/tags/$tag\") | \"\(.object.type) \(.object.sha)\"")"
  if [ -z "$matching" ]; then
    echo "missing"
    return 0
  fi
  if [ "$matching" = "commit $sha" ]; then
    echo "ok"
    return 0
  fi
  if [[ "$matching" =~ ^tag[[:space:]]+(.+)$ ]]; then
    local obj_sha="${BASH_REMATCH[1]}"
    if [[ ! "$obj_sha" =~ ^[0-9a-f]{40}$ ]]; then
      echo "error: invalid tag object sha '$obj_sha' for tag $tag" >&2
      exit 1
    fi
    local deref
    deref="$(gh_read "read tag object for $tag" api "repos/$REPO/git/tags/$obj_sha" --jq '"\(.object.type) \(.object.sha)"')"
    if [ "$deref" = "commit $sha" ]; then
      echo "ok"
      return 0
    fi
    echo "error: tag $tag exists at $deref, want commit $sha" >&2
    exit 1
  fi
  echo "error: tag $tag exists at $matching, want commit $sha" >&2
  exit 1
}

# Why: gh pr list --label reads the search index, which lags behind a merge; the REST list with a head filter does not; --paginate so a pending pull request behind 100 newer closed ones is still found.
OWNER="${REPO%%/*}"
pr_lines="$(gh_read "list release pull requests for $REPO" api --paginate "repos/$REPO/pulls?state=closed&head=$OWNER:$RELEASE_BRANCH&per_page=100" --jq '.[] | select(.merged_at != null) | select(.head.ref == "'"$RELEASE_BRANCH"'") | select(any(.labels[]; .name == "autorelease: pending")) | "\(.number) \(.merge_commit_sha)"')"

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

if [ "$MODE" = "tag" ]; then
  all_tags=()
  for pr_entry in "${prs[@]}"; do
    [[ "$pr_entry" =~ ^([1-9][0-9]*)[[:space:]]+([0-9a-f]{40})$ ]]
    pr_num="${BASH_REMATCH[1]}"
    sha="${BASH_REMATCH[2]}"

    planned_tags="$(plan_tags "$pr_num" "$sha")"
    mapfile -t tags <<< "$planned_tags"
    for tag in "${tags[@]}"; do
      state="$(tag_state "$tag" "$sha")"
      if [ "$state" = "missing" ]; then
        if ! gh_try_write POST "repos/$REPO/git/refs" -f "ref=refs/tags/$tag" -f "sha=$sha"; then
          recheck="$(tag_state "$tag" "$sha")"
          if [ "$recheck" = "ok" ]; then
            echo "tag $tag created by a concurrent run" >&2
          else
            echo "error: failed to create tag $tag at $sha" >&2
            exit 1
          fi
        fi
      elif [ "$state" = "ok" ]; then
        echo "tag $tag already exists at commit $sha" >&2
      fi
      all_tags+=("$tag")
    done
  done

  if [ ${#all_tags[@]} -eq 0 ]; then
    echo '[]'
  else
    printf '%s\n' "${all_tags[@]}" | LC_ALL=C sort -u | jq -R . | jq -sc .
  fi

elif [ "$MODE" = "relabel" ]; then
  ready_prs=()
  for pr_entry in "${prs[@]}"; do
    [[ "$pr_entry" =~ ^([1-9][0-9]*)[[:space:]]+([0-9a-f]{40})$ ]]
    pr_num="${BASH_REMATCH[1]}"
    sha="${BASH_REMATCH[2]}"

    planned_tags="$(plan_tags "$pr_num" "$sha")"
    mapfile -t tags <<< "$planned_tags"
    ready=1
    for tag in "${tags[@]}"; do
      state="$(tag_state "$tag" "$sha")"
      if [ "$state" = "missing" ]; then
        echo "PR $pr_num: tag $tag missing, left pending for the next run" >&2
        ready=0
        break
      fi
    done
    if [ "$ready" -eq 1 ]; then
      ready_prs+=("$pr_num")
    fi
  done

  for pr_num in "${ready_prs[@]}"; do
    gh_write POST "repos/$REPO/issues/$pr_num/labels" -f "labels[]=autorelease: tagged"
    if ! gh_try_write DELETE "repos/$REPO/issues/$pr_num/labels/autorelease:%20pending"; then
      has_pending="$(gh_read "read labels of PR $pr_num" api "repos/$REPO/issues/$pr_num/labels" --jq 'any(.[]; .name == "autorelease: pending")')"
      if [ "$has_pending" = "false" ]; then
        echo "PR $pr_num: pending label already removed" >&2
      else
        echo "error: failed to remove pending label from PR $pr_num" >&2
        exit 1
      fi
    fi
  done

  if [ ${#ready_prs[@]} -eq 0 ]; then
    echo '[]'
  else
    printf '%s\n' "${ready_prs[@]}" | LC_ALL=C sort -nu | jq -s -c .
  fi
fi
