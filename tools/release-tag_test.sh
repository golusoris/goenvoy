#!/usr/bin/env bash
# Planted cases for tools/release-tag.sh against a stub gh:
#   a) one pending PR, two changed paths -> 2 ref POSTs, label POST + DELETE, stdout tags
#   b) no pending PR -> [], no writes
#   c) PR head branch not release-please--branches-- -> ignored, [], no writes
#   d) tag exists at same commit -> no ref POST for it, labels swapped, tag still in stdout
#   e) tag exists at another sha -> exit 1, no label writes
#   f) manifest unchanged -> exit 1, no writes
#   g) merge commit with two parents -> exit 1, no writes
#   h) gh failure on pr list -> exit 1
#   i) RELEASE_TAG_DRY_RUN=1 -> stdout tags, zero logged writes, stderr holds dry-run: lines
#   j) changed path version does not fit module path -> exit 1, no writes
#   k) bad sha in pr list -> exit 1
#   usage: tools/release-tag_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

fail() {
  local name="$1" msg="$2"
  printf 'FAIL %s: %s\n' "$name" "$msg" >&2
  exit 1
}

# Stub gh binary
cat > "$WORK/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

case "$1" in
  pr)
    if [ "${STUB_PR_FAIL:-0}" = "1" ]; then
      exit 1
    fi
    jq_expr=""
    for ((i=1; i<=$#; i++)); do
      if [ "${!i}" = "--jq" ]; then
        next=$((i+1))
        jq_expr="${!next}"
        break
      fi
    done
    if [ -n "$jq_expr" ]; then
      jq -r "$jq_expr" "$STUB_PR_JSON"
    else
      cat "$STUB_PR_JSON"
    fi
    ;;
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
    "headRefName": "release-please--branches--main",
    "mergeCommit": {
      "oid": "$SHA_MERGE"
    }
  }
]
EOF

LOG="$WORK/log"
OUT="$WORK/out"
ERR="$WORK/err"

run_tag() {
  : > "$LOG"
  : > "$OUT"
  : > "$ERR"
  local rc=0
  (
    GH_BIN="$WORK/gh" \
    STUB_PR_JSON="$PR_JSON" \
    STUB_COMMITS_DIR="$COMMITS_DIR" \
    STUB_MANIFESTS_DIR="$MANIFESTS_DIR" \
    STUB_MATCHING_REFS_MAP="$MATCHING_REFS" \
    STUB_LOG="$LOG" \
    RELEASE_TAG_ROOT="$REPO_DIR" \
    bash "$HERE/release-tag.sh" owner/repo > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# Case a: one pending PR, two changed paths -> 2 ref POSTs, label POST + DELETE, stdout JSON
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
    "headRefName": "feature/not-release-please",
    "mergeCommit": {
      "oid": "$SHA_MERGE"
    }
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
    "headRefName": "release-please--branches--main",
    "mergeCommit": {
      "oid": "$SHA_MERGE"
    }
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

# Case h: gh failure on pr list -> exit 1
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

# Case k: bad sha in pr list -> exit 1
cat > "$PR_JSON" <<'EOF'
[
  {
    "number": 10,
    "headRefName": "release-please--branches--main",
    "mergeCommit": {
      "oid": "bad-sha"
    }
  }
]
EOF
rc=0; run_tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_k" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_k" "want 0 writes"
echo "ok   case_k"

echo "release tag: all cases pass"
