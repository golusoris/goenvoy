#!/usr/bin/env bash
# Planted cases for the release manifest module count check.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Why: testlib is resolved relative to this script at runtime.
# shellcheck source=tools/testlib.sh disable=SC1091
. "$HERE/testlib.sh"
SUT=(bash "$HERE/readme-module-count.sh")
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [ "${1:-}" = "--fixture" ]; then
  cd "$2"
  git init -q
  # Why: ensure Git tracks the fixture rather than a redirected repository.
  git_dir="$(git rev-parse --absolute-git-dir)"
  if [ "$git_dir" != "$PWD/.git" ]; then
    printf 'FAIL tracked fixture: git resolves to %s, not the fixture; an exported GIT_DIR or GIT_WORK_TREE redirected it\n' "$git_dir" >&2
    exit 1
  fi
  mkdir -p docs assets sub
  printf '{"one":"1","two":"1"}\n' > .release-please-manifest.json
  printf '2 modules\n' > README.md
  printf '2 modules\n' > docs/index.md
  printf 'banner\n1 modules\n' > assets/banner.svg
  git add .release-please-manifest.json README.md docs/index.md assets/banner.svg
  expect_refusal 'tracked stale banner' 'assets/banner.svg:2' "${SUT[@]}"
  printf '2 modules\n' > assets/banner.svg
  printf '1 modules\n' > CHANGELOG.md
  printf '1 modules\n' > sub/CHANGELOG.md
  git add CHANGELOG.md sub/CHANGELOG.md
  expect_pass 'historical counts exempt' "${SUT[@]}"
  printf '1 modules\n' > untracked.md
  expect_pass 'untracked count ignored' "${SUT[@]}"
  printf 'no count\n' > README.md
  expect_refusal 'required readme count' 'README.md: no module count found' "${SUT[@]}"
  exit 0
fi

python3 - "$BASH" "$HERE/readme-module-count_test.sh" "$WORK" <<'PYENV'
import os
import sys

environment = {key: value for key, value in os.environ.items() if not key.startswith("GIT_")}
environment["GIT_CONFIG_GLOBAL"] = os.devnull
os.execve(sys.argv[1], [sys.argv[1], sys.argv[2], "--fixture", sys.argv[3]], environment)
PYENV

printf '{"one":"1.0.0","two":"1.0.0"}\n' > "$WORK/manifest.json"
printf '2 modules\n' > "$WORK/matching.md"
printf 'heading\n1 modules and 2 modules\n' > "$WORK/stale.md"
printf 'no count\n' > "$WORK/empty.md"
printf '{"one":"1","one":"2"}\n' > "$WORK/duplicate.json"
expect_pass 'matching count' "${SUT[@]}" "$WORK/manifest.json" "$WORK/matching.md"
expect_refusal 'stale count' "$WORK/stale.md:2: says 1 modules, manifest has 2" "${SUT[@]}" "$WORK/manifest.json" "$WORK/stale.md"
expect_refusal 'missing count' "$WORK/empty.md" "${SUT[@]}" "$WORK/manifest.json" "$WORK/empty.md"
expect_refusal 'missing manifest' "$WORK/missing.json" "${SUT[@]}" "$WORK/missing.json" "$WORK/matching.md"
expect_refusal 'duplicate manifest key' 'duplicate' "${SUT[@]}" "$WORK/duplicate.json" "$WORK/matching.md"
(cd "$HERE/.." && expect_pass 'real repository' "${SUT[@]}")
