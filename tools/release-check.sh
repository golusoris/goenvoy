#!/usr/bin/env bash
# Pre-release audit for one module.
#   usage: tools/release-check.sh <module-path> vX.Y.Z
#   example: tools/release-check.sh arr/sonarr v1.4.0
set -euo pipefail

MOD="${1?module path (e.g. arr/sonarr)}"
VER="${2?version (e.g. v1.4.0)}"

if [[ ! "$VER" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
  echo "error: version must match vX.Y.Z[-prerelease], got: $VER" >&2
  exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
if [ ! -d "$REPO_ROOT/$MOD" ]; then
  echo "error: module directory not found: $MOD" >&2
  exit 1
fi
if [ ! -f "$REPO_ROOT/$MOD/go.mod" ]; then
  echo "error: no go.mod in $MOD" >&2
  exit 1
fi

BIN_DIR=$(mktemp -d)
PREV_WT=""
cleanup() {
  rm -rf "$BIN_DIR"
  if [ -n "$PREV_WT" ] && [ -d "$PREV_WT" ]; then
    git worktree remove --force "$PREV_WT" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

echo "==> building audit tools (gosec, govulncheck, apidiff)"
(
  cd "$REPO_ROOT/tools/go"
  go build -o "$BIN_DIR/gosec" github.com/securego/gosec/v2/cmd/gosec
  go build -o "$BIN_DIR/govulncheck" golang.org/x/vuln/cmd/govulncheck
  go build -o "$BIN_DIR/apidiff" golang.org/x/exp/cmd/apidiff
)

XTOOLS_VER=$(go version -m "$BIN_DIR/gosec" | grep 'x/tools' | tail -n 1 | awk '{print $3}')
if [ -z "$XTOOLS_VER" ] || ! printf '%s\n%s\n' "v0.50.0" "$XTOOLS_VER" | sort -V -C; then
  echo "error: gosec x/tools version (${XTOOLS_VER:-missing}) < v0.50.0 floor" >&2
  exit 1
fi

cd "$REPO_ROOT/$MOD"

echo "==> go build"
go build ./...

echo "==> go vet"
go vet ./...

echo "==> go test (race)"
go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...

if command -v jq >/dev/null 2>&1; then
  echo "==> coverage threshold"
  THRESHOLD=$(jq -r --arg k "$MOD" '.overrides[$k] // .default' "$REPO_ROOT/tools/coverage-thresholds.json")
  COVERAGE=$(go tool cover -func=coverage.out | awk '/^total/{gsub("%","",$3); print $3}')
  if [ -z "$COVERAGE" ]; then
    echo "   no executable statements - skipping coverage threshold"
  elif [ "$COVERAGE" = "0.0" ] && [ "$(grep -cv '^mode:' coverage.out || true)" -eq 0 ]; then
    echo "   no executable statements - skipping coverage threshold"
  else
    echo "   coverage ${COVERAGE}% (threshold ${THRESHOLD}%)"
    if awk -v c="$COVERAGE" -v t="$THRESHOLD" 'BEGIN { exit !(c+0 < t+0) }'; then
      echo "!! Coverage ${COVERAGE}% < ${THRESHOLD}% for ${MOD}"
      exit 1
    fi
  fi
else
  echo "!! jq not installed - skipping coverage threshold"
fi

if command -v golangci-lint >/dev/null 2>&1; then
  echo "==> golangci-lint"
  golangci-lint run --config "$REPO_ROOT/.golangci.yml" ./...
else
  echo "!! golangci-lint not installed — skipping"
fi

echo "==> govulncheck"
"$BIN_DIR/govulncheck" ./...

echo "==> gosec"
"$BIN_DIR/gosec" -quiet -exclude-generated ./...

echo "==> apidiff vs previous tag"
PREV=$(git tag --list "${MOD}/v*" --sort=-v:refname | head -n 1 || true)
if [ -z "$PREV" ]; then
  echo "   no previous ${MOD}/v* tag — first release, skipping apidiff"
else
  MI=$(go list -m)
  "$BIN_DIR/apidiff" -m -w "$BIN_DIR/curr.fp" "$MI"
  PREV_WT=$(mktemp -d)
  git worktree add --detach --quiet "$PREV_WT" "$PREV"
  ( cd "$PREV_WT/$MOD" && "$BIN_DIR/apidiff" -m -w "$BIN_DIR/prev.fp" "$MI" )
  if [ ! -s "$BIN_DIR/prev.fp" ] || [ ! -s "$BIN_DIR/curr.fp" ]; then
    echo "error: apidiff produced an empty fingerprint" >&2
    exit 1
  fi
  if ! INCOMPAT=$("$BIN_DIR/apidiff" -incompatible -m "$BIN_DIR/prev.fp" "$BIN_DIR/curr.fp" 2>&1); then
    echo "error: apidiff comparison failed: $INCOMPAT" >&2
    exit 1
  fi
  if [ -n "$INCOMPAT" ]; then
    echo "!! Breaking API change vs $PREV — bump major or add BREAKING CHANGE footer"
    echo "$INCOMPAT"
    exit 1
  fi
  echo "   OK: compatible vs $PREV"
fi

echo
echo "Ready to tag: git tag ${MOD}/${VER} && git push origin ${MOD}/${VER}"
