#!/usr/bin/env bash
# Planted cases for tools/release-manifest.sh: a manifest that matches its
# modules passes and yields the tag list; a major the module path does not
# carry, a malformed version, a missing module and an unknown path fail.
#   usage: tools/release-manifest_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# plant CASE PATH VERSION [MODULE]: adds PATH at VERSION to the manifest of
# CASE; without MODULE the directory gets no go.mod.
plant() {
  local root="$WORK/$1" manifest="$WORK/$1/.release-please-manifest.json"
  mkdir -p "$root/$2"
  [ -f "$manifest" ] || echo '{}' > "$manifest"
  jq --arg p "$2" --arg v "$3" '.[$p] = $v' "$manifest" > "$manifest.new"
  mv "$manifest.new" "$manifest"
  if [ $# -ge 4 ]; then
    printf 'module %s\n\ngo 1.26\n' "$4" > "$root/$2/go.mod"
  fi
}

# expect_refusal NAME NEEDLE ARGS...: the script must fail and name NEEDLE.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$(bash "$HERE/release-manifest.sh" "$@" 2>&1)"; then
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

# expect_output NAME WANT ARGS...: the script must succeed and print WANT.
expect_output() {
  local name="$1" want="$2" out
  shift 2
  out="$(bash "$HERE/release-manifest.sh" "$@")"
  if [ "$out" != "$want" ]; then
    printf 'FAIL %s: got %s, want %s\n' "$name" "$out" "$want" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

plant clean a/one 1.3.2 example.test/a/one
plant clean b/two 2.0.1 example.test/b/two/v2
expect_output clean "release manifest: 2 modules, every version matches its module path" check "$WORK/clean"
expect_output tags '["b/two/v2.0.1","a/one/v1.3.2"]' tags '["b/two","a/one"]' "$WORK/clean"
expect_output no-release '[]' tags '[]' "$WORK/clean"

plant major-ahead b/two 4.0.0 example.test/b/two/v2
expect_refusal major-ahead "needs a module path ending in /v4" check "$WORK/major-ahead"
expect_refusal major-ahead-tags "needs a module path ending in /v4" tags '["b/two"]' "$WORK/major-ahead"

plant missing-suffix b/two 2.0.0 example.test/b/two
expect_refusal missing-suffix "needs a module path ending in /v2" check "$WORK/missing-suffix"

plant stray-suffix b/two 1.4.0 example.test/b/two/v2
expect_refusal stray-suffix "does not fit the module path" check "$WORK/stray-suffix"

plant bad-version a/one 1.3 example.test/a/one
expect_refusal bad-version "is not X.Y.Z" check "$WORK/bad-version"

plant no-module a/one 1.0.0
expect_refusal no-module "no go.mod" check "$WORK/no-module"

mkdir -p "$WORK/empty" && echo '{}' > "$WORK/empty/.release-please-manifest.json"
expect_refusal empty "lists no module" check "$WORK/empty"

expect_refusal unknown-path "not in the release manifest" tags '["c/none"]' "$WORK/clean"
expect_refusal not-a-list "not a JSON array" tags '"a/one"' "$WORK/clean"
mkdir -p "$WORK/none"
expect_refusal no-manifest "no release manifest" check "$WORK/none"

echo "release manifest check: 3 accepted, 10 refused"
