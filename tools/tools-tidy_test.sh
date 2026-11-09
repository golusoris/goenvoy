#!/usr/bin/env bash
# Why: planted cases for tools/tools-tidy.sh: a tidy copy of tools/go passes, an untidy go.mod or go.sum or an import without a requirement is refused.
set -euo pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
  echo "tools-tidy_test.sh: needs bash >= 4 (mapfile); macOS: install a current bash" >&2
  exit 2
fi

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=tools/testlib.sh disable=SC1091
. "$HERE/testlib.sh"
REPO_ROOT="$(cd "$HERE/.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Why: copy every tracked tools/go file from the repository into the temporary work directory.
copy_tracked_tools_go() {
  local -a files
  local f
  mapfile -t files < <(git -C "$REPO_ROOT" ls-files -- tools/go)
  for f in "${files[@]}"; do
    mkdir -p "$WORK/$(dirname "$f")"
    cp "$REPO_ROOT/$f" "$WORK/$f"
  done
}

# Why: restore go.sum from the original tidy copy.
restore_sum() {
  cp "$WORK/go.sum.orig" "$MOD/go.sum"
}

# Why: restore go.mod from the original tidy copy.
restore_mod() {
  cp "$WORK/go.mod.orig" "$MOD/go.mod"
}

copy_tracked_tools_go
MOD="$WORK/tools/go"

(cd "$MOD" && env GOWORK=off GOFLAGS= go mod tidy)
expect_pass "tidy copy passes" env GOWORK=off GOFLAGS= bash "$HERE/tools-tidy.sh" "$MOD"

cp "$MOD/go.mod" "$WORK/go.mod.orig"
cp "$MOD/go.sum" "$WORK/go.sum.orig"

echo "example.invalid/unused v1.0.0 h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" >> "$MOD/go.sum"
expect_refusal "stale go.sum entry" "-example.invalid/unused v1.0.0" env GOWORK=off GOFLAGS= GOPROXY=off bash "$HERE/tools-tidy.sh" "$MOD"
restore_sum

awk '{ if (!done && sub(/[ \t]+\/\/ indirect$/, "")) { done=1 } print }' "$MOD/go.mod" > "$MOD/go.mod.drift"
if cmp -s "$MOD/go.mod" "$MOD/go.mod.drift"; then
  echo "FAIL go.mod drift: plant did not apply" >&2
  exit 1
fi
mv "$MOD/go.mod.drift" "$MOD/go.mod"
expect_refusal "go.mod differs from tidy" "diff current/go.mod tidy/go.mod" env GOWORK=off GOFLAGS= GOPROXY=off bash "$HERE/tools-tidy.sh" "$MOD"
restore_mod

mkdir -p "$MOD/tidyplant"
cat <<'EOF' > "$MOD/tidyplant/plant.go"
package tidyplant

import _ "example.invalid/missing/pkg"
EOF
expect_refusal "import without a requirement" "example.invalid/missing/pkg" env GOWORK=off GOFLAGS= GOPROXY=off bash "$HERE/tools-tidy.sh" "$MOD"
rm -rf "$MOD/tidyplant"
