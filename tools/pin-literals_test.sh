#!/usr/bin/env bash
# Exercise shared pin literal checks against tracked fixture files.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=/dev/null
. "$HERE/testlib.sh"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/tools" "$WORK/nested"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" init -q
cat > "$WORK/tools/tool-versions.env" <<'PINS'
TOOL_VERSION=2.47.6
IMAGE_DIGEST=sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
PINS
printf 'clean\n' > "$WORK/example.txt"
printf '2.47.6\n' > "$WORK/nested/CHANGELOG.md"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" add tools/tool-versions.env example.txt nested/CHANGELOG.md
SUT=(bash "$HERE/pin-literals.sh" "$WORK/tools/tool-versions.env" "$WORK")
expect_pass clean "${SUT[@]}"
printf 'first\nversion 2.47.6\n' > "$WORK/example.txt"
expect_refusal version 'example.txt:2: TOOL_VERSION literal duplicated' "${SUT[@]}"
printf 'abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890\n' > "$WORK/example.txt"
expect_refusal digest 'example.txt:1: IMAGE_DIGEST literal duplicated' "${SUT[@]}"
printf '12.47.60\n' > "$WORK/example.txt"
expect_pass longer_number "${SUT[@]}"
printf 'Use 2.47.6.\n' > "$WORK/example.txt"
expect_refusal sentence_end 'example.txt:1: TOOL_VERSION literal duplicated' "${SUT[@]}"
printf 'see .2.47.6 and v2.47.6, too\n' > "$WORK/example.txt"
expect_refusal punctuation_around 'example.txt:1: TOOL_VERSION literal duplicated' "${SUT[@]}"
printf '2.47.6.1 and 1.2.47.6\n' > "$WORK/example.txt"
expect_pass longer_version "${SUT[@]}"
printf 'clean\n' > "$WORK/example.txt"
printf 'TOOL_VERSION="2.47.6"\n' > "$WORK/tools/tool-versions.env"
expect_refusal quoted_value 'unsupported pin value for TOOL_VERSION' "${SUT[@]}"
printf 'TOOL_VERSION=2.47.6 # note\n' > "$WORK/tools/tool-versions.env"
expect_refusal trailing_comment 'unsupported pin value for TOOL_VERSION' "${SUT[@]}"
: > "$WORK/tools/tool-versions.env"
expect_refusal empty 'pin file' "${SUT[@]}"
expect_pass repository bash "$HERE/pin-literals.sh"
echo 'pin-literals: all planted cases pass'
