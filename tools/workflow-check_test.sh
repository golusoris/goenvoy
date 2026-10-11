#!/usr/bin/env bash
# Planted cases for tools/workflow-check.sh: a clean workflow passes; a tag pin,
# a stale or missing version comment, a broken workflow and a missing token fail.
#   usage: tools/workflow-check_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

CHECKOUT=actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1

# plant CASE DIR USES [RUNS_ON_KEY]: one workflow whose only step uses USES.
plant() {
  mkdir -p "$WORK/$1/$2/workflows"
  cat > "$WORK/$1/$2/workflows/planted.yml" <<YAML
name: planted
on:
  pull_request:
permissions:
  contents: read
jobs:
  planted:
    ${4:-runs-on}: ubuntu-26.04
    steps:
      - uses: $3
        with:
          persist-credentials: false
YAML
}

# expect_refusal CASE NEEDLE [COMMAND PREFIX...]: the check must fail and name NEEDLE.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$("$@" bash "$HERE/workflow-check.sh" "$WORK/$name" 2>&1)"; then
    printf 'FAIL %s: accepted\n%s\n' "$name" "$out" >&2
    exit 1
  fi
  case "$out" in
    *"$needle"*) echo "ok   $name: refused ($needle)" ;;
    *)
      printf 'FAIL %s: refused without "%s"\n%s\n' "$name" "$needle" "$out" >&2
      exit 1
      ;;
  esac
}

plant clean .github "$CHECKOUT # v7.0.1"
bash "$HERE/workflow-check.sh" "$WORK/clean" > /dev/null
echo "ok   clean: accepted"

plant tag-pin .github "actions/checkout@v7"
expect_refusal tag-pin "unpinned-uses"

plant stale-comment .github "$CHECKOUT # v4.0.0"
expect_refusal stale-comment "ref-version-mismatch"

plant no-comment .github "$CHECKOUT"
expect_refusal no-comment "ref-version-mismatch"

plant broken .github "$CHECKOUT # v7.0.1" runs-in
expect_refusal broken "syntax-check"

expect_refusal clean "is required" env -u GH_TOKEN -u GITHUB_TOKEN

mkdir -p "$WORK/empty"
expect_refusal empty "no workflow file"

echo "workflow check: 1 accepted, 6 refused"
