#!/usr/bin/env bash
# Planted cases for tools/release-publish.sh against a stateful stub gh:
#   a) no release -> create, upload, view assets, edit --draft=false, view isDraft, ending isDraft false
#   b) existing draft -> no create, upload, edit, view isDraft
#   c) existing published release -> upload, view assets, no edit
#   d) upload fails -> exit 1, no edit, state still draft
#   e) one asset dropped -> exit 1, no edit, stderr names asset
#   f) three arguments -> exit 2, zero calls
#   g) repository bad repo -> exit 2, zero calls
#   h) RELEASE_PUBLISH_TIMEOUT=abc -> exit 2, zero calls
#   i) missing asset file -> exit 2, zero calls
#   usage: tools/release-publish_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

fail() {
  local name="$1" msg="$2"
  printf 'FAIL %s: %s\n' "$name" "$msg" >&2
  exit 1
}

# Stateful stub gh
cat > "$WORK/gh" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

if [ -n "${STUB_CALLS:-}" ]; then
  printf '%s\n' "$*" >> "$STUB_CALLS"
fi

case "$1" in
  release)
    subcmd="$2"
    shift 2
    case "$subcmd" in
      view)
        jq_expr=""
        while [ $# -gt 0 ]; do
          case "$1" in
            --jq)
              jq_expr="$2"
              shift 2
              ;;
            *)
              shift
              ;;
          esac
        done
        if [ ! -f "$STUB_STATE_FILE" ]; then
          exit 1
        fi
        if [ -n "$jq_expr" ]; then
          jq -r "$jq_expr" "$STUB_STATE_FILE"
        else
          cat "$STUB_STATE_FILE"
        fi
        ;;
      create)
        echo '{"isDraft":true,"assets":[]}' > "$STUB_STATE_FILE"
        ;;
      upload)
        if [ "${STUB_UPLOAD_FAIL:-0}" = "1" ]; then
          exit 1
        fi
        shift # TAG
        while [ $# -gt 0 ]; do
          case "$1" in
            --repo)
              shift 2
              ;;
            --clobber)
              shift
              ;;
            *)
              asset="$1"
              shift
              basename="${asset##*/}"
              if [ -n "${STUB_UPLOAD_DROP:-}" ] && [ "$STUB_UPLOAD_DROP" = "$basename" ]; then
                continue
              fi
              jq --arg name "$basename" '.assets += [{"name": $name}]' "$STUB_STATE_FILE" > "$STUB_STATE_FILE.tmp"
              mv "$STUB_STATE_FILE.tmp" "$STUB_STATE_FILE"
              ;;
          esac
        done
        ;;
      edit)
        while [ $# -gt 0 ]; do
          case "$1" in
            --draft=false)
              jq '.isDraft = false' "$STUB_STATE_FILE" > "$STUB_STATE_FILE.tmp"
              mv "$STUB_STATE_FILE.tmp" "$STUB_STATE_FILE"
              shift
              ;;
            *)
              shift
              ;;
          esac
        done
        ;;
      *)
        echo "stub gh: unhandled release subcmd: $subcmd" >&2
        exit 64
        ;;
    esac
    ;;
  *)
    echo "stub gh: unhandled command: $1" >&2
    exit 64
    ;;
esac
STUB
chmod 0755 "$WORK/gh"

STATE_FILE="$WORK/state.json"
CALLS="$WORK/calls"
OUT="$WORK/out"
ERR="$WORK/err"

NOTES_FILE="$WORK/notes.md"
echo "# Release Notes" > "$NOTES_FILE"

ASSET1="$WORK/asset1.tar.gz"
echo "data1" > "$ASSET1"
ASSET2="$WORK/asset2.txt"
echo "data2" > "$ASSET2"

run_publish() {
  : > "$CALLS"
  : > "$OUT"
  : > "$ERR"
  local rc=0
  (
    GH_BIN="$WORK/gh" \
    STUB_CALLS="$CALLS" \
    STUB_STATE_FILE="$STATE_FILE" \
    bash "$HERE/release-publish.sh" "$@" > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# Case a: no release -> create (args hold --draft and --verify-tag, and no asset path), upload, view assets, edit --draft=false, view isDraft, ending with isDraft false
rm -f "$STATE_FILE"
rc=0; run_publish "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_a" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 6 ] || fail "case_a" "want 6 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_a" "call 1 want release view, got ${calls[0]}"
[[ "${calls[1]}" =~ ^release[[:space:]]+create ]] || fail "case_a" "call 2 want release create, got ${calls[1]}"
[[ "${calls[1]}" =~ --draft ]] || fail "case_a" "call 2 create missing --draft"
[[ "${calls[1]}" =~ --verify-tag ]] || fail "case_a" "call 2 create missing --verify-tag"
[[ ! "${calls[1]}" =~ asset ]] || fail "case_a" "call 2 create should not hold asset path"
[[ "${calls[2]}" =~ ^release[[:space:]]+upload ]] || fail "case_a" "call 3 want release upload, got ${calls[2]}"
[[ "${calls[3]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_a" "call 4 want release view assets, got ${calls[3]}"
[[ "${calls[4]}" =~ ^release[[:space:]]+edit.*--draft=false ]] || fail "case_a" "call 5 want release edit --draft=false, got ${calls[4]}"
[[ "${calls[5]}" =~ ^release[[:space:]]+view.*isDraft ]] || fail "case_a" "call 6 want release view isDraft, got ${calls[5]}"
[ "$(jq -r .isDraft "$STATE_FILE")" = "false" ] || fail "case_a" "want isDraft false in state"
echo "ok   case_a"

# Case b: existing draft: no create, then upload and edit
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_b" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 5 ] || fail "case_b" "want 5 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_b" "call 1 want release view"
[[ "${calls[1]}" =~ ^release[[:space:]]+upload ]] || fail "case_b" "call 2 want release upload"
[[ "${calls[2]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_b" "call 3 want release view assets"
[[ "${calls[3]}" =~ ^release[[:space:]]+edit.*--draft=false ]] || fail "case_b" "call 4 want release edit --draft=false"
[[ "${calls[4]}" =~ ^release[[:space:]]+view.*isDraft ]] || fail "case_b" "call 5 want release view isDraft"
if grep -q "release create" "$CALLS"; then fail "case_b" "should not call release create"; fi
[ "$(jq -r .isDraft "$STATE_FILE")" = "false" ] || fail "case_b" "want isDraft false"
echo "ok   case_b"

# Case c: existing published release: upload, no edit
echo '{"isDraft":false,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_c" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 3 ] || fail "case_c" "want 3 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_c" "call 1 want release view"
[[ "${calls[1]}" =~ ^release[[:space:]]+upload ]] || fail "case_c" "call 2 want release upload"
[[ "${calls[2]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_c" "call 3 want release view assets"
if grep -q "release edit" "$CALLS"; then fail "case_c" "should not call release edit"; fi
grep -q "release v1.0.0 already published, assets refreshed" "$ERR" || fail "case_c" "missing refreshed message in stderr"
echo "ok   case_c"

# Case d: upload fails: exit 1, no edit, state still a draft
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0
(
  STUB_UPLOAD_FAIL=1 run_publish "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_d" "exit $rc, want 1"
if grep -q "release edit" "$CALLS"; then fail "case_d" "should not call release edit on failed upload"; fi
[ "$(jq -r .isDraft "$STATE_FILE")" = "true" ] || fail "case_d" "state should still be draft"
echo "ok   case_d"

# Case e: one asset dropped: exit 1, no edit, stderr names the asset
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0
(
  STUB_UPLOAD_DROP="asset2.txt" run_publish "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_e" "exit $rc, want 1"
if grep -q "release edit" "$CALLS"; then fail "case_e" "should not call release edit when asset dropped"; fi
grep -q "asset2.txt" "$ERR" || fail "case_e" "stderr should name dropped asset: $(cat "$ERR")"
echo "ok   case_e"

# Case f: three arguments: exit 2, zero calls
rc=0; run_publish "owner/repo" "v1.0.0" "title" || rc=$?
[ "$rc" -eq 2 ] || fail "case_f" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_f" "want 0 calls"
echo "ok   case_f"

# Case g: repository bad repo: exit 2, zero calls
rc=0; run_publish "bad repo" "v1.0.0" "title" "$NOTES_FILE" "$ASSET1" || rc=$?
[ "$rc" -eq 2 ] || fail "case_g" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_g" "want 0 calls"
echo "ok   case_g"

# Case h: RELEASE_PUBLISH_TIMEOUT=abc: exit 2, zero calls
rc=0
(
  RELEASE_PUBLISH_TIMEOUT=abc run_publish "owner/repo" "v1.0.0" "title" "$NOTES_FILE" "$ASSET1"
) || rc=$?
[ "$rc" -eq 2 ] || fail "case_h" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_h" "want 0 calls"
echo "ok   case_h"

# Case i: missing asset file: exit 2, zero calls
rc=0; run_publish "owner/repo" "v1.0.0" "title" "$NOTES_FILE" "$WORK/nonexistent.tar.gz" || rc=$?
[ "$rc" -eq 2 ] || fail "case_i" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_i" "want 0 calls"
echo "ok   case_i"

echo "release publish: all cases pass"
