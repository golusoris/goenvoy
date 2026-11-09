#!/usr/bin/env bash
# Planted cases for tools/release-tag.sh against a stub gh:
#   a) one pending PR, two changed paths -> 2 ref POSTs, label POST + DELETE, stdout tags
#   b) no pending PR -> [], no writes
#   c) PR head branch not release-please--branches-- -> ignored, [], no writes
#   d) tag exists at same commit -> no ref POST for it, labels swapped, tag still in stdout
#   e) tag exists at another sha -> exit 1, no label writes
#   f) manifest unchanged -> exit 1, no writes
#   g) merge commit with two parents -> exit 1, no writes
#   h) gh failure on pulls endpoint -> exit 1
#   i) RELEASE_TAG_DRY_RUN=1 -> stdout tags, zero logged writes, stderr holds dry-run: lines
#   j) changed path version does not fit module path -> exit 1, no writes
#   k) bad sha in pull request list -> exit 1
#   l) closed pull request with merged_at null and label autorelease: pending -> [], no writes
#   m) merged release pull request labelled autorelease: tagged only -> [], no writes
#   n) RELEASE_TAG_TIMEOUT=abc -> exit 2, no gh call
#   o) stub sleeps 5 s on pulls endpoint when STUB_SLEEP=1, RELEASE_TAG_TIMEOUT=1 -> exit 1 within 4 s, stderr names step
#   p) two changed paths whose C-locale order differs from natural-language order -> stdout tags in C-locale order
#   q) REPO bad repo -> exit 2, no gh call
#   usage: tools/release-tag_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# fail NAME MSG: prints the failed case and exits 1.
fail() {
  local name="$1" msg="$2"
  printf 'FAIL %s: %s\n' "$name" "$msg" >&2
  exit 1
}

# Stub gh binary
cat > "$WORK/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

if [ -n "${STUB_CALLS:-}" ]; then
  printf '%s\n' "$*" >> "$STUB_CALLS"
fi

case "$1" in
  api)
    shift
    if [ "${1:-}" = "--method" ]; then
      method="$2"
      shift 2
      endpoint="$1"
      shift
      printf '%s %s %s\n' "$method" "$endpoint" "$*" >> "$STUB_LOG"
      exit 0
    fi

    endpoint=""
    jq_expr=""
    while [ $# -gt 0 ]; do
      case "$1" in
        -H) shift 2 ;;
        --jq)
          jq_expr="$2"
          shift 2
          ;;
        *)
          if [ -z "$endpoint" ]; then
            endpoint="$1"
          fi
          shift
          ;;
      esac
    done

    case "$endpoint" in
      repos/*/pulls\?*|repos/*/pulls*)
        if [ "${STUB_PR_FAIL:-0}" = "1" ]; then
          echo "500 Internal Server Error" >&2
          exit 1
        fi
        if [ "${STUB_SLEEP:-0}" = "1" ]; then
          sleep 5
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$STUB_PR_JSON"
        else
          cat "$STUB_PR_JSON"
        fi
        ;;
      repos/*/commits/*)
        sha="${endpoint##*/}"
        commit_file="$STUB_COMMITS_DIR/$sha.json"
        if [ ! -f "$commit_file" ]; then
          echo "404 Not Found" >&2
          exit 1
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$commit_file"
        else
          cat "$commit_file"
        fi
        ;;
      repos/*/contents/.release-please-manifest.json\?ref=*)
        ref="${endpoint##*?ref=}"
        manifest_file="$STUB_MANIFESTS_DIR/$ref.json"
        if [ ! -f "$manifest_file" ]; then
          echo "404 Not Found" >&2
          exit 1
        fi
        cat "$manifest_file"
        ;;
      repos/*/git/matching-refs/tags/*)
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$STUB_MATCHING_REFS_MAP"
        else
          cat "$STUB_MATCHING_REFS_MAP"
        fi
        ;;
      *)
        echo "stub gh: unhandled endpoint: $endpoint" >&2
        exit 64
        ;;
    esac
    ;;
  *)
    exit 64
    ;;
esac
STUB
chmod 0755 "$WORK/gh"

# RELEASE_TAG_ROOT structure
REPO_DIR="$WORK/repo"
mkdir -p "$REPO_DIR/mod1" "$REPO_DIR/mod2" "$REPO_DIR/modv2/v2"
cat > "$REPO_DIR/mod1/go.mod" <<'EOF'
module github.com/golusoris/goenvoy/mod1

go 1.26
EOF

cat > "$REPO_DIR/mod2/go.mod" <<'EOF'
module github.com/golusoris/goenvoy/mod2

go 1.26
EOF

cat > "$REPO_DIR/modv2/v2/go.mod" <<'EOF'
module github.com/golusoris/goenvoy/modv2/v2

go 1.26
EOF

cat > "$REPO_DIR/.release-please-manifest.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
}
EOF

COMMITS_DIR="$WORK/commits"
MANIFESTS_DIR="$WORK/manifests"
mkdir -p "$COMMITS_DIR" "$MANIFESTS_DIR"

SHA_MERGE="1111111111111111111111111111111111111111"
SHA_PARENT="2222222222222222222222222222222222222222"
SHA_PARENT2="3333333333333333333333333333333333333333"
SHA_OTHER="9999999999999999999999999999999999999999"

# Common commit setup
cat > "$COMMITS_DIR/$SHA_MERGE.json" <<EOF
{
  "parents": [
    { "sha": "$SHA_PARENT" }
  ]
}
EOF

cat > "$MANIFESTS_DIR/$SHA_PARENT.json" <<'EOF'
{
  "mod1": "1.0.0",
  "mod2": "1.0.0",
  "modv2/v2": "2.0.0"
}
EOF

cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
}
EOF

MATCHING_REFS="$WORK/matching_refs.json"
echo '[]' > "$MATCHING_REFS"

PR_JSON="$WORK/prs.json"
cat > "$PR_JSON" <<EOF
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF

LOG="$WORK/log"
OUT="$WORK/out"
ERR="$WORK/err"
CALLS="$WORK/calls"

# run_tag [REPO]: runs tools/release-tag.sh against the stub gh and captures stdout, stderr, writes and calls.
run_tag() {
  local target_repo="${1:-owner/repo}"
  : > "$LOG"
  : > "$OUT"
  : > "$ERR"
  : > "$CALLS"
  local rc=0
  (
    GH_BIN="$WORK/gh" \
    STUB_CALLS="$CALLS" \
    STUB_PR_JSON="$PR_JSON" \
    STUB_COMMITS_DIR="$COMMITS_DIR" \
    STUB_MANIFESTS_DIR="$MANIFESTS_DIR" \
    STUB_MATCHING_REFS_MAP="$MATCHING_REFS" \
    STUB_LOG="$LOG" \
    RELEASE_TAG_ROOT="$REPO_DIR" \
    bash "$HERE/release-tag.sh" "$target_repo" > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# Case a: one pending PR, two changed paths -> 2 ref POSTs, label POST + DELETE, stdout tags
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_a" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_a" "stdout $(cat "$OUT")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 2 ] || fail "case_a" "want 2 ref POSTs"
grep -q "POST repos/owner/repo/git/refs -f ref=refs/tags/mod1/v1.1.0 -f sha=$SHA_MERGE" "$LOG" || fail "case_a" "missing mod1 ref POST"
grep -q "POST repos/owner/repo/git/refs -f ref=refs/tags/mod2/v1.0.1 -f sha=$SHA_MERGE" "$LOG" || fail "case_a" "missing mod2 ref POST"
grep -q 'POST repos/owner/repo/issues/10/labels -f labels\[\]=autorelease: tagged' "$LOG" || fail "case_a" "missing label POST"
grep -q 'DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending' "$LOG" || fail "case_a" "missing label DELETE"
echo "ok   case_a"

# Case b: no pending PR -> [], no writes
echo '[]' > "$PR_JSON"
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_b" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_b" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_b" "want 0 writes"
echo "ok   case_b"

# Case c: PR head branch not release-please--branches-- -> ignored, [], no writes
cat > "$PR_JSON" <<EOF
[
  {
    "number": 11,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "feature/not-release-please"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_c" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_c" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_c" "want 0 writes"
echo "ok   case_c"

# Reset PR_JSON for subsequent cases
cat > "$PR_JSON" <<EOF
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF

# Case d: tag exists at same commit -> no ref POST for it, labels swapped, tag still in stdout
cat > "$MATCHING_REFS" <<EOF
[
  {
    "ref": "refs/tags/mod1/v1.1.0",
    "object": {
      "type": "commit",
      "sha": "$SHA_MERGE"
    }
  },
  {
    "ref": "refs/tags/mod2/v1.0.1",
    "object": {
      "type": "commit",
      "sha": "$SHA_MERGE"
    }
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_d" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_d" "stdout $(cat "$OUT")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 0 ] || fail "case_d" "want 0 ref POSTs"
grep -q 'POST repos/owner/repo/issues/10/labels -f labels\[\]=autorelease: tagged' "$LOG" || fail "case_d" "missing label POST"
grep -q 'DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending' "$LOG" || fail "case_d" "missing label DELETE"
echo "ok   case_d"

# Case e: tag exists at another sha -> exit 1, no label writes
cat > "$MATCHING_REFS" <<EOF
[
  {
    "ref": "refs/tags/mod1/v1.1.0",
    "object": {
      "type": "commit",
      "sha": "$SHA_OTHER"
    }
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_e" "exit $rc, want 1"
grep -q "error: tag mod1/v1.1.0 exists at commit $SHA_OTHER, want commit $SHA_MERGE" "$ERR" || fail "case_e" "unexpected error: $(cat "$ERR")"
[ "$(grep -c 'labels' "$LOG")" -eq 0 ] || fail "case_e" "unexpected label writes"
echo '[]' > "$MATCHING_REFS"
echo "ok   case_e"

# Case f: manifest unchanged -> exit 1, no writes
cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "mod1": "1.0.0",
  "mod2": "1.0.0",
  "modv2/v2": "2.0.0"
}
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_f" "exit $rc, want 1"
grep -q "error: release pull request 10 changes no manifest version" "$ERR" || fail "case_f" "unexpected error: $(cat "$ERR")"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_f" "want 0 writes"
# Restore merge manifest
cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
}
EOF
echo "ok   case_f"

# Case g: merge commit with two parents -> exit 1, no writes
cat > "$COMMITS_DIR/$SHA_MERGE.json" <<EOF
{
  "parents": [
    { "sha": "$SHA_PARENT" },
    { "sha": "$SHA_PARENT2" }
  ]
}
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_g" "exit $rc, want 1"
grep -q "error: release pull request 10 merge commit $SHA_MERGE must have exactly one parent" "$ERR" || fail "case_g" "unexpected error: $(cat "$ERR")"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_g" "want 0 writes"
# Restore commit
cat > "$COMMITS_DIR/$SHA_MERGE.json" <<EOF
{
  "parents": [
    { "sha": "$SHA_PARENT" }
  ]
}
EOF
echo "ok   case_g"

# Case h: gh failure on pulls endpoint -> exit 1
rc=0
(
  STUB_PR_FAIL=1 run_tag
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_h" "exit $rc, want 1"
echo "ok   case_h"

# Case i: RELEASE_TAG_DRY_RUN=1 -> stdout tags, zero logged writes, stderr holds dry-run: lines
rc=0
(
  RELEASE_TAG_DRY_RUN=1 run_tag
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_i" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_i" "stdout $(cat "$OUT")"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_i" "want 0 writes in dry-run"
grep -q "dry-run: POST repos/owner/repo/git/refs ref=refs/tags/mod1/v1.1.0 sha=$SHA_MERGE" "$ERR" || fail "case_i" "missing mod1 dry-run ref"
grep -q "dry-run: POST repos/owner/repo/git/refs ref=refs/tags/mod2/v1.0.1 sha=$SHA_MERGE" "$ERR" || fail "case_i" "missing mod2 dry-run ref"
grep -q "dry-run: POST repos/owner/repo/issues/10/labels labels\[\]=autorelease: tagged" "$ERR" || fail "case_i" "missing dry-run label POST"
grep -q "dry-run: DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending" "$ERR" || fail "case_i" "missing dry-run label DELETE"
echo "ok   case_i"

# Case j: changed path whose version does not fit its module path (2.0.0 for a path whose go.mod module has no /v2) -> exit 1, no writes
cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "mod1": "2.0.0",
  "mod2": "1.0.0",
  "modv2/v2": "2.0.0"
}
EOF
cat > "$REPO_DIR/.release-please-manifest.json" <<'EOF'
{
  "mod1": "2.0.0",
  "mod2": "1.0.0",
  "modv2/v2": "2.0.0"
}
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_j" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_j" "want 0 writes"
# Restore manifests
cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
}
EOF
cat > "$REPO_DIR/.release-please-manifest.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
}
EOF
echo "ok   case_j"

# Case k: bad sha in pull request list -> exit 1
cat > "$PR_JSON" <<'EOF'
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "bad-sha",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_k" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_k" "want 0 writes"
echo "ok   case_k"

# Case l: closed pull request with merged_at null and label autorelease: pending -> [], no writes
cat > "$PR_JSON" <<EOF
[
  {
    "number": 73,
    "merged_at": null,
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_l" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_l" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_l" "want 0 writes"
echo "ok   case_l"

# Case m: merged release pull request labelled autorelease: tagged only -> [], no writes
cat > "$PR_JSON" <<EOF
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: tagged" }
    ]
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_m" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_m" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_m" "want 0 writes"
echo "ok   case_m"

# Case n: RELEASE_TAG_TIMEOUT=abc -> exit 2, no gh call
rc=0
(
  RELEASE_TAG_TIMEOUT=abc run_tag
) || rc=$?
[ "$rc" -eq 2 ] || fail "case_n" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_n" "want 0 gh calls, got $(wc -l < "$CALLS")"
echo "ok   case_n"

# Case o: stub sleeps 5 s on pulls endpoint when STUB_SLEEP=1, RELEASE_TAG_TIMEOUT=1 -> exit 1 within 4 s, stderr names step
cat > "$PR_JSON" <<EOF
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF
start_sec=$SECONDS
rc=0
(
  STUB_SLEEP=1 RELEASE_TAG_TIMEOUT=1 run_tag
) || rc=$?
elapsed=$((SECONDS - start_sec))
[ "$rc" -eq 1 ] || fail "case_o" "exit $rc, want 1"
[ "$elapsed" -lt 4 ] || fail "case_o" "elapsed $elapsed s, want < 4 s"
grep -q "error: failed to list release pull requests for owner/repo" "$ERR" || fail "case_o" "stderr did not name the step: $(cat "$ERR")"
echo "ok   case_o"

# Case p: two changed paths whose C-locale order differs from natural-language order -> stdout tags in C-locale order
mkdir -p "$REPO_DIR/metadata/game/steam" "$REPO_DIR/metadata/game/steamgriddb"
cat > "$REPO_DIR/metadata/game/steam/go.mod" <<'EOF'
module github.com/golusoris/goenvoy/metadata/game/steam

go 1.26
EOF

cat > "$REPO_DIR/metadata/game/steamgriddb/go.mod" <<'EOF'
module github.com/golusoris/goenvoy/metadata/game/steamgriddb

go 1.26
EOF

cat > "$REPO_DIR/.release-please-manifest.json" <<'EOF'
{
  "metadata/game/steam": "1.3.6",
  "metadata/game/steamgriddb": "1.0.7"
}
EOF

cat > "$MANIFESTS_DIR/$SHA_PARENT.json" <<'EOF'
{
  "metadata/game/steam": "1.3.5",
  "metadata/game/steamgriddb": "1.0.6"
}
EOF

cat > "$MANIFESTS_DIR/$SHA_MERGE.json" <<'EOF'
{
  "metadata/game/steam": "1.3.6",
  "metadata/game/steamgriddb": "1.0.7"
}
EOF

cat > "$PR_JSON" <<EOF
[
  {
    "number": 10,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: pending" }
    ]
  }
]
EOF

echo '[]' > "$MATCHING_REFS"

# Why: sort without LC_ALL=C puts steamgriddb first under en_US or de_DE collation, so the case runs under one when installed.
natural_locale=""
if all_locales="$(locale -a 2>/dev/null)" && match="$(grep -i -m1 -E '^(en_US|de_DE)\.utf-?8$' <<< "$all_locales")"; then
  natural_locale="$match"
fi
rc=0
(
  LC_ALL="${natural_locale:-C}" run_tag
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_p" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["metadata/game/steam/v1.3.6","metadata/game/steamgriddb/v1.0.7"]' ] || fail "case_p" "stdout $(cat "$OUT"), want [\"metadata/game/steam/v1.3.6\",\"metadata/game/steamgriddb/v1.0.7\"]"
echo "ok   case_p (locale ${natural_locale:-C only, no natural-language locale installed})"

# Case q: REPO bad repo -> exit 2, no gh call
rc=0; run_tag "bad repo" || rc=$?
[ "$rc" -eq 2 ] || fail "case_q" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_q" "want 0 gh calls, got $(wc -l < "$CALLS")"
echo "ok   case_q"

echo "release tag: all cases pass"
