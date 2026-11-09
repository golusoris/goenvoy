#!/usr/bin/env bash
# Planted cases for tools/release-publish.sh against a stateful stub gh:
#   a) stage, no release: create (--draft, --verify-tag, no asset basename in its arguments), upload, view assets, in that order. No edit. The state ends as isDraft true with both assets
#   b) stage, existing draft: upload, view assets. No create, no edit
#   c) stage, existing published release: upload, view assets, no edit, stderr has already published
#   d) stage, upload fails: exit 1, no edit, state still a draft
#   e) stage, one asset dropped: exit 1, no edit, stderr names the asset
#   f) stage, three arguments after the mode: exit 2, zero calls
#   g) stage, repository bad repo: exit 2, zero calls
#   h) stage, RELEASE_PUBLISH_TIMEOUT=abc: exit 2, zero calls
#   i) stage, missing asset file: exit 2, zero calls
#   j) stage with STUB_VIEW_ERROR='HTTP 502: Bad Gateway' and no state file: exit 1, zero create calls, stderr has HTTP 502
#   k) confirm on a published state: rc 0, one view call
#   l) confirm on a draft state: exit 1, stderr has still a draft
#   m) confirm with STUB_VIEW_ERROR set: exit 1
#   n) no mode argument, and mode publish: exit 2, zero calls (two runs)
#   o) the release.yml step order: find line numbers of stage, edit --draft=false, confirm in release.yml; exactly once and in that order inside 'Create or update GitHub Release'
#   usage: tools/release-publish_test.sh
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
        if [ -n "${STUB_VIEW_ERROR:-}" ]; then
          echo "$STUB_VIEW_ERROR" >&2
          exit 1
        fi
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
          echo "release not found" >&2
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

# run_publish ARGS...: runs tools/release-publish.sh against the stub gh and captures stdout, stderr and calls.
run_publish() {
  : > "$CALLS"
  : > "$OUT"
  : > "$ERR"
  local rc=0
  (
    GH_BIN="$WORK/gh" \
    STUB_CALLS="$CALLS" \
    STUB_STATE_FILE="$STATE_FILE" \
    STUB_VIEW_ERROR="${STUB_VIEW_ERROR:-}" \
    bash "$HERE/release-publish.sh" "$@" > "$OUT" 2> "$ERR"
  ) || rc=$?
  return "$rc"
}

# Case a: stage, no release: create (--draft, --verify-tag, no asset basename in its arguments), upload, view assets, in that order. No edit. The state ends as isDraft true with both assets
rm -f "$STATE_FILE"
rc=0; run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_a" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 4 ] || fail "case_a" "want 4 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_a" "call 1 want release view, got ${calls[0]}"
[[ "${calls[1]}" =~ ^release[[:space:]]+create ]] || fail "case_a" "call 2 want release create, got ${calls[1]}"
[[ "${calls[1]}" =~ --draft ]] || fail "case_a" "call 2 create missing --draft"
[[ "${calls[1]}" =~ --verify-tag ]] || fail "case_a" "call 2 create missing --verify-tag"
[[ ! "${calls[1]}" =~ asset ]] || fail "case_a" "call 2 create should not hold asset path"
[[ "${calls[2]}" =~ ^release[[:space:]]+upload ]] || fail "case_a" "call 3 want release upload, got ${calls[2]}"
[[ "${calls[3]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_a" "call 4 want release view assets, got ${calls[3]}"
if grep -q "release edit" "$CALLS"; then fail "case_a" "should not call release edit"; fi
[ "$(jq -r .isDraft "$STATE_FILE")" = "true" ] || fail "case_a" "want isDraft true in state"
[ "$(jq '.assets | length' "$STATE_FILE")" -eq 2 ] || fail "case_a" "want 2 assets in state"
grep -q "release v1.0.0 staged as a draft, assets verified" "$ERR" || fail "case_a" "missing staged message in stderr"
echo "ok   case_a"

# Case b: stage, existing draft: upload, view assets. No create, no edit
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_b" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 3 ] || fail "case_b" "want 3 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_b" "call 1 want release view"
[[ "${calls[1]}" =~ ^release[[:space:]]+upload ]] || fail "case_b" "call 2 want release upload"
[[ "${calls[2]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_b" "call 3 want release view assets"
if grep -q "release create" "$CALLS"; then fail "case_b" "should not call release create"; fi
if grep -q "release edit" "$CALLS"; then fail "case_b" "should not call release edit"; fi
[ "$(jq -r .isDraft "$STATE_FILE")" = "true" ] || fail "case_b" "want isDraft true"
echo "ok   case_b"

# Case c: stage, existing published release: upload, view assets, no edit, stderr has already published
echo '{"isDraft":false,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2" || rc=$?
[ "$rc" -eq 0 ] || fail "case_c" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 3 ] || fail "case_c" "want 3 calls, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_c" "call 1 want release view"
[[ "${calls[1]}" =~ ^release[[:space:]]+upload ]] || fail "case_c" "call 2 want release upload"
[[ "${calls[2]}" =~ ^release[[:space:]]+view.*assets ]] || fail "case_c" "call 3 want release view assets"
if grep -q "release create" "$CALLS"; then fail "case_c" "should not call release create"; fi
if grep -q "release edit" "$CALLS"; then fail "case_c" "should not call release edit"; fi
grep -q "already published" "$ERR" || fail "case_c" "missing already published in stderr"
echo "ok   case_c"

# Case d: stage, upload fails: exit 1, no edit, state still a draft
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0
(
  STUB_UPLOAD_FAIL=1 run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_d" "exit $rc, want 1"
if grep -q "release edit" "$CALLS"; then fail "case_d" "should not call release edit on failed upload"; fi
[ "$(jq -r .isDraft "$STATE_FILE")" = "true" ] || fail "case_d" "state should still be draft"
echo "ok   case_d"

# Case e: stage, one asset dropped: exit 1, no edit, stderr names the asset
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0
(
  STUB_UPLOAD_DROP="asset2.txt" run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_e" "exit $rc, want 1"
if grep -q "release edit" "$CALLS"; then fail "case_e" "should not call release edit when asset dropped"; fi
grep -q "asset2.txt" "$ERR" || fail "case_e" "stderr should name dropped asset: $(cat "$ERR")"
echo "ok   case_e"

# Case f: stage, three arguments after the mode: exit 2, zero calls
rc=0; run_publish stage "owner/repo" "v1.0.0" "title" || rc=$?
[ "$rc" -eq 2 ] || fail "case_f" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_f" "want 0 calls"
echo "ok   case_f"

# Case g: stage, repository bad repo: exit 2, zero calls
rc=0; run_publish stage "bad repo" "v1.0.0" "title" "$NOTES_FILE" "$ASSET1" || rc=$?
[ "$rc" -eq 2 ] || fail "case_g" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_g" "want 0 calls"
echo "ok   case_g"

# Case h: stage, RELEASE_PUBLISH_TIMEOUT=abc: exit 2, zero calls
rc=0
(
  RELEASE_PUBLISH_TIMEOUT=abc run_publish stage "owner/repo" "v1.0.0" "title" "$NOTES_FILE" "$ASSET1"
) || rc=$?
[ "$rc" -eq 2 ] || fail "case_h" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_h" "want 0 calls"
echo "ok   case_h"

# Case i: stage, missing asset file: exit 2, zero calls
rc=0; run_publish stage "owner/repo" "v1.0.0" "title" "$NOTES_FILE" "$WORK/nonexistent.tar.gz" || rc=$?
[ "$rc" -eq 2 ] || fail "case_i" "exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_i" "want 0 calls"
echo "ok   case_i"

# Case j: stage with STUB_VIEW_ERROR='HTTP 502: Bad Gateway' and no state file: exit 1, zero create calls, stderr has HTTP 502
rm -f "$STATE_FILE"
rc=0
(
  STUB_VIEW_ERROR='HTTP 502: Bad Gateway' run_publish stage "owner/repo" "v1.0.0" "mod: v1.0.0" "$NOTES_FILE" "$ASSET1" "$ASSET2"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_j" "exit $rc, want 1"
if grep -q "release create" "$CALLS"; then fail "case_j" "should not call release create on view error"; fi
grep -q "HTTP 502" "$ERR" || fail "case_j" "stderr missing HTTP 502: $(cat "$ERR")"
echo "ok   case_j"

# Case k: confirm on a published state: rc 0, one view call
echo '{"isDraft":false,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish confirm "owner/repo" "v1.0.0" || rc=$?
[ "$rc" -eq 0 ] || fail "case_k" "exit $rc, want 0"
[ "$(wc -l < "$CALLS")" -eq 1 ] || fail "case_k" "want 1 call, got $(wc -l < "$CALLS")"
mapfile -t calls < "$CALLS"
[[ "${calls[0]}" =~ ^release[[:space:]]+view ]] || fail "case_k" "call should be release view, got ${calls[0]}"
grep -q "release v1.0.0 published" "$ERR" || fail "case_k" "missing published message in stderr: $(cat "$ERR")"
echo "ok   case_k"

# Case l: confirm on a draft state: exit 1, stderr has still a draft
echo '{"isDraft":true,"assets":[]}' > "$STATE_FILE"
rc=0; run_publish confirm "owner/repo" "v1.0.0" || rc=$?
[ "$rc" -eq 1 ] || fail "case_l" "exit $rc, want 1"
grep -q "still a draft" "$ERR" || fail "case_l" "stderr missing 'still a draft': $(cat "$ERR")"
echo "ok   case_l"

# Case m: confirm with STUB_VIEW_ERROR set: exit 1
rc=0
(
  STUB_VIEW_ERROR='HTTP 502: Bad Gateway' run_publish confirm "owner/repo" "v1.0.0"
) || rc=$?
[ "$rc" -eq 1 ] || fail "case_m" "exit $rc, want 1"
echo "ok   case_m"

# Case n: no mode argument, and mode publish: exit 2, zero calls (two runs)
rc=0; run_publish || rc=$?
[ "$rc" -eq 2 ] || fail "case_n" "no mode: exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_n" "no mode: want 0 calls"

rc=0; run_publish publish "owner/repo" "v1.0.0" || rc=$?
[ "$rc" -eq 2 ] || fail "case_n" "mode publish: exit $rc, want 2"
[ "$(wc -l < "$CALLS")" -eq 0 ] || fail "case_n" "mode publish: want 0 calls"
echo "ok   case_n"

# Case o: the release.yml step order
WF_FILE="$HERE/../.github/workflows/release.yml"
[ -f "$WF_FILE" ] || fail "case_o" "missing $WF_FILE"

step_line="$(grep -n 'name:[[:space:]]*Create or update GitHub Release' "$WF_FILE" | cut -d: -f1)"
[ -n "$step_line" ] || fail "case_o" "cannot find step 'Create or update GitHub Release'"

next_step_offset=""
if tail -n +"$((step_line + 1))" "$WF_FILE" | grep -q '^[[:space:]]*- name:'; then
  next_step_offset="$(tail -n +"$((step_line + 1))" "$WF_FILE" | grep -n '^[[:space:]]*- name:' | head -n 1 | cut -d: -f1)"
fi

if [ -n "$next_step_offset" ]; then
  step_end="$((step_line + next_step_offset))"
else
  step_end="$(( $(wc -l < "$WF_FILE") + 1 ))"
fi

mapfile -t stage_lines < <(grep -n 'release-publish\.sh[[:space:]]\+stage' "$WF_FILE" | cut -d: -f1)
mapfile -t edit_lines < <(grep -F -n "gh release edit \"\$TAG\" --repo \"\$GITHUB_REPOSITORY\" --draft=false" "$WF_FILE" | cut -d: -f1)
mapfile -t confirm_lines < <(grep -n 'release-publish\.sh[[:space:]]\+confirm' "$WF_FILE" | cut -d: -f1)

[ "${#stage_lines[@]}" -eq 1 ] || fail "case_o" "stage occurs ${#stage_lines[@]} times, want 1"
[ "${#edit_lines[@]}" -eq 1 ] || fail "case_o" "edit occurs ${#edit_lines[@]} times, want 1"
[ "${#confirm_lines[@]}" -eq 1 ] || fail "case_o" "confirm occurs ${#confirm_lines[@]} times, want 1"

l_stage="${stage_lines[0]}"
l_edit="${edit_lines[0]}"
l_confirm="${confirm_lines[0]}"

if [ "$l_stage" -le "$step_line" ] || [ "$l_stage" -ge "$step_end" ]; then
  fail "case_o" "stage at line $l_stage outside step ($step_line..$step_end)"
fi
if [ "$l_edit" -le "$step_line" ] || [ "$l_edit" -ge "$step_end" ]; then
  fail "case_o" "edit at line $l_edit outside step ($step_line..$step_end)"
fi
if [ "$l_confirm" -le "$step_line" ] || [ "$l_confirm" -ge "$step_end" ]; then
  fail "case_o" "confirm at line $l_confirm outside step ($step_line..$step_end)"
fi

if [ "$l_stage" -ge "$l_edit" ] || [ "$l_edit" -ge "$l_confirm" ]; then
  fail "case_o" "lines not in order: stage=$l_stage, edit=$l_edit, confirm=$l_confirm"
fi

echo "ok   case_o"

echo "release publish: all cases pass"
