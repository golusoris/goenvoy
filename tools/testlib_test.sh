#!/usr/bin/env bash
# Exercise the test library: a fixture repository stays apart from the repository a hook exports.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/caller" "$WORK/fixture"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK/caller" init -q
printf 'probe\n' > "$WORK/fixture/probe.txt"
# Why: the library is sourced with the caller's repository exported, the way a hook runs a test.
GIT_DIR="$WORK/caller/.git" GIT_INDEX_FILE="$WORK/caller/.git/index" GIT_WORK_TREE="$WORK/caller" \
  GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 bash -c '
    set -euo pipefail
    . "$1/testlib.sh"
    git -C "$2" init -q
    git -C "$2" add probe.txt
  ' _ "$HERE" "$WORK/fixture"
# shellcheck source=tools/testlib.sh
. "$HERE/testlib.sh"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
expect_output caller_index_untouched '' git -C "$WORK/caller" ls-files
expect_output fixture_owns_file 'probe.txt' git -C "$WORK/fixture" ls-files
echo "testlib: all planted cases pass"
