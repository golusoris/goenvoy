#!/usr/bin/env bash
# Lint every workflow file and refuse an action that is not pinned to a commit
# or whose version comment does not name that commit.
#   usage: tools/workflow-check.sh [ROOT]   (default: this repository)
#   needs: docker, jq, and GH_TOKEN or GITHUB_TOKEN (read access is enough)
# Interim gate: it calls actionlint and zizmor directly until the shared
# workflow check can be called from a public repository.
set -euo pipefail

if [ $# -gt 1 ]; then
  echo "usage: ${0##*/} [ROOT]" >&2
  exit 2
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT="$(cd "${1:-$REPO_ROOT}" && pwd)"
DOCKER="${DOCKER_BIN:-docker}"

# zizmor compares a version comment with the upstream tags, which it reads
# only with a token; without one it drops that audit and still exits 0.
TOKEN="${GH_TOKEN:-${GITHUB_TOKEN:-}}"
if [ -z "$TOKEN" ]; then
  echo "error: GH_TOKEN or GITHUB_TOKEN is required: without a token the version comments are not checked" >&2
  exit 1
fi

# shellcheck source=/dev/null
. "$REPO_ROOT/tools/tool-versions.env"
# mirror.gcr.io serves Docker Hub's digests; hosted runners share Docker Hub's anonymous pull limit.
ACTIONLINT_IMAGE="mirror.gcr.io/rhysd/actionlint:${ACTIONLINT_VERSION}@${ACTIONLINT_IMAGE_DIGEST}"
ZIZMOR_IMAGE="ghcr.io/zizmorcore/zizmor:${ZIZMOR_VERSION}@${ZIZMOR_IMAGE_DIGEST}"

shopt -s nullglob
FILES=()
for f in "$ROOT"/.github/workflows/*.yml "$ROOT"/.github/workflows/*.yaml \
         "$ROOT"/.gitea/workflows/*.yml "$ROOT"/.gitea/workflows/*.yaml; do
  FILES+=("${f#"$ROOT"/}")
done
if [ ${#FILES[@]} -eq 0 ]; then
  echo "error: no workflow file under $ROOT" >&2
  exit 1
fi

"$DOCKER" run --rm -v "$ROOT:/repo:ro" -w /repo \
  -v "$REPO_ROOT/.github/actionlint.yaml:/actionlint.yaml:ro" \
  "$ACTIONLINT_IMAGE" -no-color -config-file /actionlint.yaml "${FILES[@]}"
echo "actionlint: ${#FILES[@]} workflow files passed"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
rc=0
GH_TOKEN="$TOKEN" "$DOCKER" run --rm -e GH_TOKEN -v "$ROOT:/repo:ro" -w /repo \
  "$ZIZMOR_IMAGE" --no-progress --color never --format json --persona pedantic \
  --no-ignores --no-config --strict-collection "${FILES[@]}" \
  > "$WORK/report.json" 2> "$WORK/zizmor.log" || rc=$?
# 0 = no finding, 11 to 14 = findings by highest severity; anything else is a failed run.
case "$rc" in
  0 | 11 | 12 | 13 | 14) ;;
  *)
    cat "$WORK/zizmor.log" >&2
    echo "error: zizmor exit $rc" >&2
    exit 1
    ;;
esac

# Only the two pin audits decide; zizmor's other audits are not adopted here yet.
FINDINGS="$(jq -r '
  .[]
  | select(.ident == "unpinned-uses" or .ident == "ref-version-mismatch")
  | (.locations | map(select(.symbolic.kind == "Primary")) | .[0]) as $at
  | "\($at.symbolic.key.Local.verbatim_path):\($at.concrete.location.start_point.row + 1): \(.ident): \($at.symbolic.annotation)"
' "$WORK/report.json")"
if [ -n "$FINDINGS" ]; then
  printf '%s\n' "$FINDINGS" >&2
  echo "error: pin every action to a commit and keep its version comment on that commit's tag" >&2
  exit 1
fi
echo "zizmor: every action in ${#FILES[@]} workflow files is pinned to a commit with a matching version comment"
