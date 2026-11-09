#!/usr/bin/env bash
# Planted cases for tools/release-tag.sh against a stub gh:
#   a) one pending PR, two changed paths -> 2 ref POSTs, no label writes, stdout tags
#   b) no pending PR -> [], no writes
#   c) PR head branch not release-please--branches-- -> ignored, [], no writes
#   d) tag exists at same commit -> no ref POST for it, no label writes, tag still in stdout
#   e) tag exists at another sha -> exit 1, no writes
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
#   r) tag: 101 pulls entries, 100 merged and labelled tagged, with pending PR last -> stdout tags, pulls call carries --paginate
#   s) tag: annotated tag object at merge commit -> no ref POST, tag in stdout
#   t) tag: annotated tag object at another commit -> exit 1, no writes
#   u) tag: ref POST fails, recheck sees commit SHA -> rc 0, tag in stdout, stderr has created by a concurrent run
#   v) tag: ref POST fails, recheck still missing -> exit 1, stderr has failed to create tag
#   w) relabel: all tags present at SHA -> label POST, then DELETE, no ref write, stdout [<N>]
#   x) relabel: one tag missing -> no writes, stdout [], stderr names PR and missing tag
#   y1) relabel: DELETE fails, labels read has no pending -> rc 0, stderr has already removed
#   y2) relabel: DELETE fails, pending still listed -> exit 1
#   z1) no mode argument -> exit 2, zero gh calls
#   z2) mode publish -> exit 2, zero gh calls
#   i2) relabel with RELEASE_TAG_DRY_RUN=1 -> stdout [<N>], zero logged writes, dry-run: lines in stderr
#   aa) tag with STUB_MATCHING_FAIL=1 -> exit 1, zero logged writes, stderr has failed to check matching refs, stdout empty
#   ab) relabel with STUB_MATCHING_FAIL=1 -> exit 1, zero logged writes, stderr does not have left pending
#   ac) tag with STUB_COMMITS_FAIL=1 -> exit 1, zero logged writes, stderr has failed to fetch commit
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

# Sort guard: stub sort refuses invocation without LC_ALL=C
REAL_SORT="$(command -v sort)"
mkdir -p "$WORK/bin"
cat > "$WORK/bin/sort" <<EOF
#!/usr/bin/env bash
if [ "\${LC_ALL:-}" != "C" ]; then
  echo "stub sort: called without LC_ALL=C" >&2
  exit 64
fi
exec "$REAL_SORT" "\$@"
EOF
chmod 0755 "$WORK/bin/sort"

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
      if [ "$method" = "POST" ] && [[ "$endpoint" =~ repos/.+/git/refs ]] && [ "${STUB_REF_POST_FAIL:-0}" = "1" ]; then
        if [ -n "${STUB_WORK:-}" ]; then
          touch "$STUB_WORK/ref_post_failed"
        fi
        exit 1
      fi
      if [ "$method" = "DELETE" ] && [[ "$endpoint" =~ repos/.+/issues/[0-9]+/labels/ ]] && [ "${STUB_LABEL_DELETE_FAIL:-0}" = "1" ]; then
        exit 1
      fi
      exit 0
    fi

    paginate=0
    endpoint=""
    jq_expr=""
    while [ $# -gt 0 ]; do
      case "$1" in
        --paginate)
          paginate=1
          shift
          ;;
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
        pr_source="$STUB_PR_JSON"
        if [ "$paginate" -eq 0 ]; then
          per_page=30
          if [[ "$endpoint" =~ per_page=([0-9]+) ]]; then
            per_page="${BASH_REMATCH[1]}"
          fi
          pr_source="$STUB_WORK/prs_paginated.json"
          jq --argjson n "$per_page" '.[0:$n]' "$STUB_PR_JSON" > "$pr_source"
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$pr_source"
        else
          cat "$pr_source"
        fi
        ;;
      repos/*/commits/*)
        if [ "${STUB_COMMITS_FAIL:-0}" = "1" ]; then
          echo "HTTP 502" >&2
          exit 1
        fi
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
        if [ "${STUB_MATCHING_FAIL:-0}" = "1" ]; then
          echo "HTTP 502" >&2
          exit 1
        fi
        refs_file="$STUB_MATCHING_REFS_MAP"
        if [ -n "${STUB_WORK:-}" ] && [ -f "$STUB_WORK/ref_post_failed" ] && [ -n "${STUB_REFS_AFTER_FAIL:-}" ] && [ -f "$STUB_REFS_AFTER_FAIL" ]; then
          refs_file="$STUB_REFS_AFTER_FAIL"
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$refs_file"
        else
          cat "$refs_file"
        fi
        ;;
      repos/*/git/tags/*)
        tag_sha="${endpoint##*/}"
        line=""
        if [ -n "${STUB_TAG_OBJECTS_MAP:-}" ] && [ -f "$STUB_TAG_OBJECTS_MAP" ]; then
          if grep -q "^$tag_sha " "$STUB_TAG_OBJECTS_MAP"; then
            line="$(grep "^$tag_sha " "$STUB_TAG_OBJECTS_MAP")"
          fi
        fi
        if [ -z "$line" ]; then
          echo "404 Not Found" >&2
          exit 1
        fi
        read -r _ t_type t_target <<< "$line"
        json="{\"object\":{\"type\":\"$t_type\",\"sha\":\"$t_target\"}}"
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" <<< "$json"
        else
          echo "$json"
        fi
        ;;
      repos/*/issues/*/labels)
        labels_file="${STUB_LABELS_JSON:-}"
        if [ -z "$labels_file" ] || [ ! -f "$labels_file" ]; then
          echo "404 Not Found" >&2
          exit 1
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$labels_file"
        else
          cat "$labels_file"
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

TAG_OBJECTS_MAP="$WORK/tag_objects.txt"
: > "$TAG_OBJECTS_MAP"

LABELS_JSON="$WORK/labels.json"
echo '[]' > "$LABELS_JSON"

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

# run_tag MODE [REPO]: runs tools/release-tag.sh against the stub gh and captures stdout, stderr, writes and calls.
run_tag() {
  local mode="$1"
  local target_repo="${2:-owner/repo}"
  : > "$LOG"
  : > "$OUT"
  : > "$ERR"
  : > "$CALLS"
  rm -f "$WORK/ref_post_failed"
  local rc=0
  (
    PATH="$WORK/bin:$PATH" \
    GH_BIN="$WORK/gh" \
    STUB_WORK="$WORK" \
    STUB_CALLS="$CALLS" \
    STUB_PR_JSON="$PR_JSON" \
    STUB_COMMITS_DIR="$COMMITS_DIR" \
    STUB_MANIFESTS_DIR="$MANIFESTS_DIR" \
    STUB_MATCHING_REFS_MAP="$MATCHING_REFS" \
    STUB_TAG_OBJECTS_MAP="$TAG_OBJECTS_MAP" \
    STUB_LABELS_JSON="$LABELS_JSON" \
    STUB_LOG="$LOG" \
    RELEASE_TAG_ROOT="$REPO_DIR" \
    bash "$HERE/release-tag.sh" "$mode" "$target_repo" > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# run_tag_no_args: runs tools/release-tag.sh without arguments.
run_tag_no_args() {
  : > "$LOG"
  : > "$OUT"
  : > "$ERR"
  : > "$CALLS"
  local rc=0
  (
    PATH="$WORK/bin:$PATH" \
    GH_BIN="$WORK/gh" \
    STUB_WORK="$WORK" \
    STUB_CALLS="$CALLS" \
    STUB_PR_JSON="$PR_JSON" \
    STUB_COMMITS_DIR="$COMMITS_DIR" \
    STUB_MANIFESTS_DIR="$MANIFESTS_DIR" \
    STUB_MATCHING_REFS_MAP="$MATCHING_REFS" \
    STUB_TAG_OBJECTS_MAP="$TAG_OBJECTS_MAP" \
    STUB_LABELS_JSON="$LABELS_JSON" \
    STUB_LOG="$LOG" \
    RELEASE_TAG_ROOT="$REPO_DIR" \
    bash "$HERE/release-tag.sh" > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# Case a: one pending PR, two changed paths -> 2 ref POSTs, no label writes, stdout tags
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_a" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_a" "stdout $(cat "$OUT")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 2 ] || fail "case_a" "want 2 ref POSTs"
grep -q "POST repos/owner/repo/git/refs -f ref=refs/tags/mod1/v1.1.0 -f sha=$SHA_MERGE" "$LOG" || fail "case_a" "missing mod1 ref POST"
grep -q "POST repos/owner/repo/git/refs -f ref=refs/tags/mod2/v1.0.1 -f sha=$SHA_MERGE" "$LOG" || fail "case_a" "missing mod2 ref POST"
[ "$(grep -c 'labels' "$LOG")" -eq 0 ] || fail "case_a" "unexpected label writes"
echo "ok   case_a"

# Case b: no pending PR -> [], no writes
echo '[]' > "$PR_JSON"
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
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

# Case d: tag exists at same commit -> no ref POST for it, no label writes, tag still in stdout
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
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_d" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_d" "stdout $(cat "$OUT")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 0 ] || fail "case_d" "want 0 ref POSTs"
[ "$(grep -c 'labels' "$LOG")" -eq 0 ] || fail "case_d" "unexpected label writes"
echo "ok   case_d"

# Case e: tag exists at another sha -> exit 1, no writes
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
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
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
  STUB_PR_FAIL=1 run_tag tag
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_h" "exit $rc, want 1"
echo "ok   case_h"

# Case i: RELEASE_TAG_DRY_RUN=1 -> stdout tags, zero logged writes, stderr holds dry-run: lines
rc=0
(
  RELEASE_TAG_DRY_RUN=1 run_tag tag
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_i" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_i" "stdout $(cat "$OUT")"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_i" "want 0 writes in dry-run"
grep -q "dry-run: POST repos/owner/repo/git/refs ref=refs/tags/mod1/v1.1.0 sha=$SHA_MERGE" "$ERR" || fail "case_i" "missing mod1 dry-run ref"
grep -q "dry-run: POST repos/owner/repo/git/refs ref=refs/tags/mod2/v1.0.1 sha=$SHA_MERGE" "$ERR" || fail "case_i" "missing mod2 dry-run ref"
if grep -q "labels" "$ERR"; then
  fail "case_i" "unexpected dry-run label writes"
fi
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
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
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
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_m" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_m" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_m" "want 0 writes"
echo "ok   case_m"

# Case n: RELEASE_TAG_TIMEOUT=abc -> exit 2, no gh call
rc=0
(
  RELEASE_TAG_TIMEOUT=abc run_tag tag
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
  STUB_SLEEP=1 RELEASE_TAG_TIMEOUT=1 run_tag tag
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

# Why: the stub sort refuses any sort without LC_ALL=C, and the probed locale also shows the natural order.
natural_locale=""
if all_locales="$(locale -a 2>/dev/null)" && match="$(grep -i -m1 -E '^(en_US|de_DE)\.utf-?8$' <<< "$all_locales")"; then
  natural_locale="$match"
fi
rc=0
(
  LC_ALL="${natural_locale:-}" run_tag tag
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_p" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["metadata/game/steam/v1.3.6","metadata/game/steamgriddb/v1.0.7"]' ] || fail "case_p" "stdout $(cat "$OUT"), want [\"metadata/game/steam/v1.3.6\",\"metadata/game/steamgriddb/v1.0.7\"]"
echo "ok   case_p (locale ${natural_locale:-C only, no natural-language locale installed})"

# Case q: REPO bad repo -> exit 2, no gh call
rc=0; run_tag tag "bad repo" || rc=$?
[ "$rc" -eq 2 ] || fail "case_q" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_q" "want 0 gh calls, got $(wc -l < "$CALLS")"
echo "ok   case_q"

# Reset fixtures for new cases
cat > "$REPO_DIR/.release-please-manifest.json" <<'EOF'
{
  "mod1": "1.1.0",
  "mod2": "1.0.1",
  "modv2/v2": "2.0.0"
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

# Case r: tag: 101 pulls entries, 100 merged and labelled tagged, with pending PR last -> stdout tags, pulls call carries --paginate
cat > "$PR_JSON" <<EOF
[
EOF
for i in $(seq 1 100); do
  cat >> "$PR_JSON" <<EOF
  {
    "number": $i,
    "merged_at": "2026-11-09T12:00:00Z",
    "merge_commit_sha": "$SHA_MERGE",
    "head": {
      "ref": "release-please--branches--main"
    },
    "labels": [
      { "name": "autorelease: tagged" }
    ]
  },
EOF
done
cat >> "$PR_JSON" <<EOF
  {
    "number": 101,
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
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_r" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_r" "stdout $(cat "$OUT")"
grep -q -- '--paginate' "$CALLS" || fail "case_r" "missing --paginate in pulls call"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 2 ] || fail "case_r" "want 2 ref POSTs"
echo "ok   case_r"

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

# Case s: tag: annotated tag object at merge commit -> no ref POST, tag in stdout
SHA_TAG_OBJ1="4444444444444444444444444444444444444444"
SHA_TAG_OBJ2="5555555555555555555555555555555555555555"
cat > "$MATCHING_REFS" <<EOF
[
  {
    "ref": "refs/tags/mod1/v1.1.0",
    "object": {
      "type": "tag",
      "sha": "$SHA_TAG_OBJ1"
    }
  },
  {
    "ref": "refs/tags/mod2/v1.0.1",
    "object": {
      "type": "tag",
      "sha": "$SHA_TAG_OBJ2"
    }
  }
]
EOF
cat > "$TAG_OBJECTS_MAP" <<EOF
$SHA_TAG_OBJ1 commit $SHA_MERGE
$SHA_TAG_OBJ2 commit $SHA_MERGE
EOF
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 0 ] || fail "case_s" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_s" "stdout $(cat "$OUT")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -eq 0 ] || fail "case_s" "want 0 ref POSTs"
echo "ok   case_s"

# Case t: tag: annotated tag object at another commit -> exit 1, no writes
cat > "$TAG_OBJECTS_MAP" <<EOF
$SHA_TAG_OBJ1 commit $SHA_OTHER
$SHA_TAG_OBJ2 commit $SHA_MERGE
EOF
rc=0; run_tag tag || rc=$?
[ "$rc" -eq 1 ] || fail "case_t" "exit $rc, want 1"
grep -q "error: tag mod1/v1.1.0 exists at commit $SHA_OTHER, want commit $SHA_MERGE" "$ERR" || fail "case_t" "unexpected error: $(cat "$ERR")"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_t" "want 0 writes"
echo "ok   case_t"

# Case u: tag: ref POST fails, recheck sees commit SHA -> rc 0, tag in stdout, stderr has created by a concurrent run
echo '[]' > "$MATCHING_REFS"
REFS_AFTER="$WORK/refs_after_fail.json"
cat > "$REFS_AFTER" <<EOF
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
rc=0
(
  STUB_REF_POST_FAIL=1 STUB_REFS_AFTER_FAIL="$REFS_AFTER" run_tag tag
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_u" "exit $rc, want 0"
[ "$(cat "$OUT")" = '["mod1/v1.1.0","mod2/v1.0.1"]' ] || fail "case_u" "stdout $(cat "$OUT")"
grep -q "tag mod1/v1.1.0 created by a concurrent run" "$ERR" || fail "case_u" "missing concurrent run message in stderr: $(cat "$ERR")"
[ "$(grep -c 'POST repos/owner/repo/git/refs' "$LOG")" -ge 1 ] || fail "case_u" "want logged ref POST"
echo "ok   case_u"

# Case v: tag: ref POST fails, recheck still missing -> exit 1, stderr has failed to create tag
echo '[]' > "$MATCHING_REFS"
rc=0
(
  STUB_REF_POST_FAIL=1 run_tag tag
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_v" "exit $rc, want 1"
grep -q "error: failed to create tag mod1/v1.1.0 at $SHA_MERGE" "$ERR" || fail "case_v" "unexpected error in stderr: $(cat "$ERR")"
echo "ok   case_v"

# Case w: relabel: all tags present at SHA -> label POST, then DELETE, no ref write, stdout [<N>]
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
rc=0; run_tag relabel || rc=$?
[ "$rc" -eq 0 ] || fail "case_w" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[10]' ] || fail "case_w" "stdout $(cat "$OUT"), want [10]"
[ "$(grep -c 'git/refs' "$LOG")" -eq 0 ] || fail "case_w" "want 0 ref writes"
grep -q 'POST repos/owner/repo/issues/10/labels -f labels\[\]=autorelease: tagged' "$LOG" || fail "case_w" "missing label POST"
grep -q 'DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending' "$LOG" || fail "case_w" "missing label DELETE"
echo "ok   case_w"

# Case x: relabel: one tag missing -> no writes, stdout [], stderr names PR and missing tag
cat > "$MATCHING_REFS" <<EOF
[
  {
    "ref": "refs/tags/mod1/v1.1.0",
    "object": {
      "type": "commit",
      "sha": "$SHA_MERGE"
    }
  }
]
EOF
rc=0; run_tag relabel || rc=$?
[ "$rc" -eq 0 ] || fail "case_x" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[]' ] || fail "case_x" "stdout $(cat "$OUT"), want []"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_x" "want 0 writes"
grep -q "PR 10: tag mod2/v1.0.1 missing, left pending for the next run" "$ERR" || fail "case_x" "unexpected stderr: $(cat "$ERR")"
echo "ok   case_x"

# Case y1: relabel: DELETE fails, labels read has no pending -> rc 0, stderr has already removed
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
cat > "$LABELS_JSON" <<EOF
[
  { "name": "autorelease: tagged" }
]
EOF
rc=0
(
  STUB_LABEL_DELETE_FAIL=1 run_tag relabel
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_y1" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[10]' ] || fail "case_y1" "stdout $(cat "$OUT"), want [10]"
grep -q "PR 10: pending label already removed" "$ERR" || fail "case_y1" "unexpected stderr: $(cat "$ERR")"
grep -q 'POST repos/owner/repo/issues/10/labels -f labels\[\]=autorelease: tagged' "$LOG" || fail "case_y1" "missing label POST"
grep -q 'DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending' "$LOG" || fail "case_y1" "missing label DELETE"
echo "ok   case_y1"

# Case y2: relabel: DELETE fails, pending still listed -> exit 1
cat > "$LABELS_JSON" <<EOF
[
  { "name": "autorelease: pending" },
  { "name": "autorelease: tagged" }
]
EOF
rc=0
(
  STUB_LABEL_DELETE_FAIL=1 run_tag relabel
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_y2" "exit $rc, want 1"
echo "ok   case_y2"

# Case z1: no mode argument -> exit 2, zero gh calls
rc=0; run_tag_no_args || rc=$?
[ "$rc" -eq 2 ] || fail "case_z1" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_z1" "want 0 gh calls, got $(wc -l < "$CALLS")"
echo "ok   case_z1"

# Case z2: mode publish -> exit 2, zero gh calls
rc=0; run_tag publish || rc=$?
[ "$rc" -eq 2 ] || fail "case_z2" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_z2" "want 0 gh calls, got $(wc -l < "$CALLS")"
echo "ok   case_z2"

# Case i2: relabel with RELEASE_TAG_DRY_RUN=1 -> stdout [<N>], zero logged writes, dry-run: lines in stderr
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
rc=0
(
  RELEASE_TAG_DRY_RUN=1 run_tag relabel
) || rc=$?
[ "$rc" -eq 0 ] || fail "case_i2" "exit $rc, want 0"
[ "$(cat "$OUT")" = '[10]' ] || fail "case_i2" "stdout $(cat "$OUT"), want [10]"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_i2" "want 0 writes in dry-run"
grep -q "dry-run: POST repos/owner/repo/issues/10/labels labels\[\]=autorelease: tagged" "$ERR" || fail "case_i2" "missing dry-run label POST"
grep -q "dry-run: DELETE repos/owner/repo/issues/10/labels/autorelease:%20pending" "$ERR" || fail "case_i2" "missing dry-run label DELETE"
echo "ok   case_i2"

# Case aa: tag with STUB_MATCHING_FAIL=1 -> exit 1, zero logged writes, stderr has failed to check matching refs, stdout empty
rc=0
(
  STUB_MATCHING_FAIL=1 run_tag tag
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_aa" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_aa" "want 0 writes"
[ ! -s "$OUT" ] || fail "case_aa" "stdout not empty: $(cat "$OUT")"
grep -q "failed to check matching refs" "$ERR" || fail "case_aa" "stderr missing 'failed to check matching refs': $(cat "$ERR")"
echo "ok   case_aa"

# Case ab: relabel with STUB_MATCHING_FAIL=1 -> exit 1, zero logged writes, stderr does not have left pending
rc=0
(
  STUB_MATCHING_FAIL=1 run_tag relabel
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_ab" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_ab" "want 0 writes"
if grep -q "left pending" "$ERR"; then
  fail "case_ab" "stderr has 'left pending': $(cat "$ERR")"
fi
echo "ok   case_ab"

# Case ac: tag with STUB_COMMITS_FAIL=1 -> exit 1, zero logged writes, stderr has failed to fetch commit
rc=0
(
  STUB_COMMITS_FAIL=1 run_tag tag
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_ac" "exit $rc, want 1"
[ "$(wc -l < "$LOG")" -eq 0 ] || fail "case_ac" "want 0 writes"
grep -q "failed to fetch commit" "$ERR" || fail "case_ac" "stderr missing 'failed to fetch commit': $(cat "$ERR")"
echo "ok   case_ac"

echo "release tag: all cases pass"
