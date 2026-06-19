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

if command -v govulncheck >/dev/null 2>&1; then
  echo "==> govulncheck"
  govulncheck ./...
else
  echo "!! govulncheck not installed — skipping (go install golang.org/x/vuln/cmd/govulncheck@latest)"
fi

if command -v gosec >/dev/null 2>&1; then
  echo "==> gosec"
  gosec -quiet -exclude-generated ./...
else
  echo "!! gosec not installed — skipping (go install github.com/securego/gosec/v2/cmd/gosec@latest)"
fi

echo "==> apidiff vs previous tag"
PREV=$(git tag --list "${MOD}/v*" --sort=-v:refname | head -1 || true)
if [ -z "$PREV" ]; then
  echo "   no previous ${MOD}/v* tag — first release, skipping apidiff"
elif ! command -v apidiff >/dev/null 2>&1; then
  echo "!! apidiff not installed — skipping (go install golang.org/x/exp/cmd/apidiff@latest)"
else
  MI=$(go list -m)
  apidiff -m "$MI" . > /tmp/release-check-curr.txt 2>/dev/null || true
  PREV_WT=$(mktemp -d)
  git worktree add --detach --quiet "$PREV_WT" "$PREV"
  trap 'git worktree remove --force "$PREV_WT" >/dev/null 2>&1 || true' EXIT
  apidiff -m "$MI" "$PREV_WT/$MOD" > /tmp/release-check-prev.txt 2>/dev/null || true
  if apidiff /tmp/release-check-prev.txt /tmp/release-check-curr.txt; then
    echo "   OK: compatible vs $PREV"
  else
    echo "!! Breaking API change vs $PREV — bump major or add BREAKING CHANGE footer"
    exit 1
  fi
fi

echo
echo "Ready to tag: git tag ${MOD}/${VER} && git push origin ${MOD}/${VER}"
