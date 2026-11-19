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
