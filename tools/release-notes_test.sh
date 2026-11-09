#!/usr/bin/env bash
# Planted cases for tools/release-notes.sh: first section, middle section
# stops before next heading, unlinked heading, prefix collision refusal,
# missing version refusal, missing file refusal, and real changelog section extraction.
#   usage: tools/release-notes_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$HERE/.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# fail NAME MSG: prints the failed case and exits 1.
fail() {
  local name="$1" msg="$2"
  printf 'FAIL %s: %s\n' "$name" "$msg" >&2
  exit 1
}

# Case 1: first section
cat > "$WORK/c1.md" <<'EOF'
# Changelog

## [2.0.0](https://example.com/2.0.0) (2026-02-01)

### Features
* new feature 1


## [1.0.0](https://example.com/1.0.0) (2026-01-01)

### Features
* initial release
EOF

rc=0
bash "$HERE/release-notes.sh" "$WORK/c1.md" "2.0.0" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 0 ] || fail "first-section" "exit $rc, want 0"
grep -q '^## \[2\.0\.0\]' "$WORK/out" || fail "first-section" "missing 2.0.0 header"
grep -q '\* new feature 1' "$WORK/out" || fail "first-section" "missing body"
if grep -q '## \[1\.0\.0\]' "$WORK/out"; then
  fail "first-section" "included subsequent section"
fi
# Trailing blank lines dropped: last line must not be empty
last_line="$(tail -n 1 "$WORK/out")"
[ -n "${last_line//[[:space:]]/}" ] || fail "first-section" "has trailing blank line"
echo "ok   first-section"

# Case 2: middle section stops before next ##
cat > "$WORK/c2.md" <<'EOF'
# Changelog

## [3.0.0](https://example.com/3.0.0) (2026-03-01)

* v3 notes

## [2.0.0](https://example.com/2.0.0) (2026-02-01)

### Bug Fixes
* fixed issue 42

## [1.0.0](https://example.com/1.0.0) (2026-01-01)

* initial release
EOF

rc=0
bash "$HERE/release-notes.sh" "$WORK/c2.md" "2.0.0" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 0 ] || fail "middle-section" "exit $rc, want 0"
grep -q '^## \[2\.0\.0\]' "$WORK/out" || fail "middle-section" "missing 2.0.0 header"
grep -q '\* fixed issue 42' "$WORK/out" || fail "middle-section" "missing body"
if grep -q '3\.0\.0' "$WORK/out" || grep -q '1\.0\.0' "$WORK/out"; then
  fail "middle-section" "included lines from outside middle section"
fi
last_line="$(tail -n 1 "$WORK/out")"
[ -n "${last_line//[[:space:]]/}" ] || fail "middle-section" "has trailing blank line"
echo "ok   middle-section"

# Case 3: heading without link ## 1.0.0 (2026-01-01)
cat > "$WORK/c3.md" <<'EOF'
# Changelog

## 1.0.0 (2026-01-01)

### Features
* unlinked header note
EOF

rc=0
bash "$HERE/release-notes.sh" "$WORK/c3.md" "1.0.0" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 0 ] || fail "heading-without-link" "exit $rc, want 0"
grep -q '^## 1\.0\.0 (2026-01-01)' "$WORK/out" || fail "heading-without-link" "missing unlinked header"
grep -q '\* unlinked header note' "$WORK/out" || fail "heading-without-link" "missing body"
echo "ok   heading-without-link"

# Case 4: 1.3.7 does not match a ## [1.3.70] section (exit 1)
cat > "$WORK/c4.md" <<'EOF'
# Changelog

## [1.3.70](https://example.com) (2026-01-01)

* should not match 1.3.7
EOF

rc=0
bash "$HERE/release-notes.sh" "$WORK/c4.md" "1.3.7" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || fail "prefix-collision" "exit $rc, want 1"
grep -q 'error:.*no section for 1\.3\.7' "$WORK/out" || fail "prefix-collision" "unexpected error message: $(cat "$WORK/out")"
echo "ok   prefix-collision"

# Case 5: missing version exit 1
rc=0
bash "$HERE/release-notes.sh" "$WORK/c1.md" "9.9.9" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 1 ] || fail "missing-version" "exit $rc, want 1"
grep -q 'error:.*no section for 9\.9\.9' "$WORK/out" || fail "missing-version" "unexpected error message: $(cat "$WORK/out")"
echo "ok   missing-version"

# Case 6: missing file exit 2
rc=0
bash "$HERE/release-notes.sh" "$WORK/does-not-exist.md" "1.0.0" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 2 ] || fail "missing-file" "exit $rc, want 2"
echo "ok   missing-file"

# Case 7: real changelog metadata/tracking/trakt/CHANGELOG.md version 1.3.7
rc=0
bash "$HERE/release-notes.sh" "$REPO_ROOT/metadata/tracking/trakt/CHANGELOG.md" "1.3.7" > "$WORK/out" 2>&1 || rc=$?
[ "$rc" -eq 0 ] || fail "real-changelog" "exit $rc, want 0"
first_line="$(head -n 1 "$WORK/out")"
case "$first_line" in
  "## [1.3.7]("* ) ;;
  * ) fail "real-changelog" "output does not start with ## [1.3.7](, got: $first_line" ;;
esac
if grep -q '^## \[1\.3\.6\]' "$WORK/out"; then
  fail "real-changelog" "output contains line starting ## [1.3.6]"
fi
echo "ok   real-changelog"

echo "release notes: all cases pass"
