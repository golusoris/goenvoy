#!/usr/bin/env bash
# Planted cases for tools/release-pr-ci.sh against a stub gh: the waiting runs
# of a release branch head are approved once each; a run that already ran, a
# run of another branch or repository, and no release pull request approve
# nothing; a head without a run, a failing or odd API answer fail the script.
#   usage: tools/release-pr-ci_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

SHA=0123456789abcdef0123456789abcdef01234567

# The stub answers from STUB_HEADS and STUB_RUNS and logs every approval.
cat > "$WORK/gh" <<'STUB'
#!/usr/bin/env bash
case "$1" in
  pr) printf '%s' "$STUB_HEADS" ;;
  api)
    if [ "$2" = --method ]; then
      [ "$STUB_APPROVE" = fail ] && exit 1
      printf '%s\n' "$4" >> "$STUB_LOG"
      exit 0
    fi
    [ "$STUB_RUNS" = fail ] && exit 1
    printf '%s' "$STUB_RUNS"
    ;;
  *) exit 64 ;;
esac
STUB
chmod 0755 "$WORK/gh"

# run_case NAME HEADS RUNS WANT_RC WANT_APPROVALS [APPROVE] [AUTHORS]
run_case() {
  local name="$1" rc=0 got
  : > "$WORK/log"
  (
    [ $# -ge 7 ] && export RELEASE_PR_AUTHORS="$7"
    GH_BIN="$WORK/gh" STUB_HEADS="$2" STUB_RUNS="$3" STUB_LOG="$WORK/log" \
      STUB_APPROVE="${6:-ok}" RELEASE_PR_CI_TRIES=2 RELEASE_PR_CI_WAIT=0 \
      bash "$HERE/release-pr-ci.sh" owner/repo > "$WORK/out" 2>&1
  ) || rc=$?
  got="$(wc -l < "$WORK/log")"
  if [ "$rc" -ne "$4" ] || [ "$got" -ne "$5" ]; then
    printf 'FAIL %s: exit %s with %s approvals, want exit %s with %s\n' "$name" "$rc" "$got" "$4" "$5" >&2
    cat "$WORK/out" >&2
    exit 1
  fi
  echo "ok   $name: exit $rc, $got approvals"
}

HEAD="app/plurano-release release-branch $SHA"

run_case waiting "$HEAD" "11 action_required release-branch owner/repo
12 action_required release-branch owner/repo" 0 2
grep -qx 'repos/owner/repo/actions/runs/11/approve' "$WORK/log"
grep -qx 'repos/owner/repo/actions/runs/12/approve' "$WORK/log"
run_case one-of-two "$HEAD" "11 success release-branch owner/repo
12 action_required release-branch owner/repo" 0 1
grep -qx 'repos/owner/repo/actions/runs/12/approve' "$WORK/log"
run_case already-ran "$HEAD" "11 success release-branch owner/repo
12 none release-branch owner/repo" 0 0
run_case other-branch "$HEAD" "11 action_required other-branch owner/repo" 0 0
run_case other-repository "$HEAD" "11 action_required release-branch fork/repo" 0 0
run_case no-release-pr "" "" 0 0
run_case no-run "$HEAD" "" 1 0
grep -q 'no pull request run on release-branch' "$WORK/out"
run_case api-fails "$HEAD" fail 1 0
run_case approve-fails "$HEAD" "11 action_required release-branch owner/repo" 1 0 fail
run_case odd-run "$HEAD" "null action_required release-branch owner/repo" 1 0
run_case odd-head "app/plurano-release release-branch not-a-sha" "11 action_required release-branch owner/repo" 1 0
run_case github-actions "app/github-actions release-branch $SHA" "11 action_required release-branch owner/repo" 0 1
grep -qx 'repos/owner/repo/actions/runs/11/approve' "$WORK/log"
run_case wrong-author "app/evil-bot release-branch $SHA" "11 action_required release-branch owner/repo" 0 0
grep -q 'Refused release pull request' "$WORK/out"
run_case human-author "someone release-branch $SHA" "11 action_required release-branch owner/repo" 0 0
grep -q 'Refused release pull request' "$WORK/out"
run_case mixed "app/plurano-release release-branch $SHA
app/evil-bot other-branch $SHA" "11 action_required release-branch owner/repo
12 action_required other-branch owner/repo" 0 1
grep -qx 'repos/owner/repo/actions/runs/11/approve' "$WORK/log"
grep -q 'Refused release pull request' "$WORK/out"
run_case bad-authors "$HEAD" "" 2 0 ok "app/x;rm"
grep -q 'error: unexpected RELEASE_PR_AUTHORS' "$WORK/out"

echo "release pull request CI: 5 approved, 4 skipped, 8 refused"
