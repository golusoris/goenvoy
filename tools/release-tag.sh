#!/usr/bin/env bash
# Tag the merged release pull requests and relabel them for publication.
#   usage: tools/release-tag.sh OWNER/REPO
#   needs: gh, jq
#   env:   GH_BIN (default: gh)
#          RELEASE_TAG_ROOT (default: repository root)
#          RELEASE_TAG_DRY_RUN (default: 0)
set -euo pipefail

usage() {
  echo "usage: ${0##*/} OWNER/REPO" >&2
  exit 2
}

[ $# -eq 1 ] || usage
REPO="$1"
[ -n "$REPO" ] || usage

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GH="${GH_BIN:-gh}"
ROOT="${RELEASE_TAG_ROOT:-$(cd "$HERE/.." && pwd)}"
DRY_RUN="${RELEASE_TAG_DRY_RUN:-0}"

for cmd in "$GH" jq; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "error: $cmd is required" >&2
    exit 1
  fi
done

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
  "$GH" api --method "$method" "$endpoint" "$@" >/dev/null || {
    echo "error: failed to $method $endpoint" >&2
    exit 1
  }
}

pr_lines="$("$GH" pr list --repo "$REPO" --state merged --label 'autorelease: pending' --limit 20 --json number,headRefName,mergeCommit --jq '.[] | select(.headRefName | startswith("release-please--branches--")) | "\(.number) \(.mergeCommit.oid)"')" || {
  echo "error: failed to list release pull requests for $REPO" >&2
  exit 1
}

prs=()
if [ -n "$pr_lines" ]; then
  while IFS= read -r pr_line || [ -n "$pr_line" ]; do
    [ -n "$pr_line" ] || continue
    if [[ ! "$pr_line" =~ ^([1-9][0-9]*)[[:space:]]+([0-9a-f]{40})$ ]]; then
      echo "error: invalid pr list entry: '$pr_line'" >&2
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

  parents="$("$GH" api "repos/$REPO/commits/$sha" --jq '.parents | map(.sha) | join(" ")')" || {
    echo "error: failed to fetch commit $sha for PR $pr_num" >&2
    exit 1
  }
  if [[ ! "$parents" =~ ^[0-9a-f]{40}$ ]]; then
    echo "error: release pull request $pr_num merge commit $sha must have exactly one parent, got: '$parents'" >&2
    exit 1
  fi
  parent="$parents"

  parent_manifest="$("$GH" api -H 'Accept: application/vnd.github.raw' "repos/$REPO/contents/.release-please-manifest.json?ref=$parent")" || {
    echo "error: failed to fetch parent manifest for PR $pr_num at $parent" >&2
    exit 1
  }
  sha_manifest="$("$GH" api -H 'Accept: application/vnd.github.raw' "repos/$REPO/contents/.release-please-manifest.json?ref=$sha")" || {
    echo "error: failed to fetch manifest for PR $pr_num at $sha" >&2
    exit 1
  }

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
    matching="$("$GH" api "repos/$REPO/git/matching-refs/tags/$tag" --jq ".[] | select(.ref == \"refs/tags/$tag\") | \"\(.object.type) \(.object.sha)\"")" || {
      echo "error: failed to check matching refs for tag $tag" >&2
      exit 1
    }
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
  printf '%s\n' "${all_tags[@]}" | sort -u | jq -R . | jq -sc .
fi
