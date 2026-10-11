#!/usr/bin/env bash
# Planted cases for tools/release-pr-ci.sh against a stub gh: a release branch
# head without a run is dispatched once; a head that has a run, and no release
# pull request, dispatch nothing; a failing or odd API answer fails the script.
#   usage: tools/release-pr-ci_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

SHA=0123456789abcdef0123456789abcdef01234567

# The stub answers from STUB_HEADS and STUB_RUNS and logs every dispatch.
cat > "$WORK/gh" <<'STUB'
#!/usr/bin/env bash
case "$1" in
  pr) printf '%s' "$STUB_HEADS" ;;
  api)
    [ "$STUB_RUNS" = fail ] && exit 1
    printf '%s\n' "$STUB_RUNS"
    ;;
  workflow) printf '%s\n' "$*" >> "$STUB_LOG" ;;
  *) exit 64 ;;
esac
STUB
chmod 0755 "$WORK/gh"

# run_case NAME HEADS RUNS WANT_RC WANT_DISPATCHES
run_case() {
  local name="$1" rc=0 got
  : > "$WORK/log"
  GH_BIN="$WORK/gh" STUB_HEADS="$2" STUB_RUNS="$3" STUB_LOG="$WORK/log" \
    bash "$HERE/release-pr-ci.sh" owner/repo > "$WORK/out" 2>&1 || rc=$?
  got="$(wc -l < "$WORK/log")"
  if [ "$rc" -ne "$4" ] || [ "$got" -ne "$5" ]; then
    printf 'FAIL %s: exit %s with %s dispatches, want exit %s with %s\n' "$name" "$rc" "$got" "$4" "$5" >&2
    cat "$WORK/out" >&2
    exit 1
  fi
  echo "ok   $name: exit $rc, $got dispatches"
}

run_case no-run "release-branch $SHA" 0 0 1
grep -qx 'workflow run ci.yml --repo owner/repo --ref release-branch' "$WORK/log"
run_case has-run "release-branch $SHA" 2 0 0
run_case no-release-pr "" 0 0 0
run_case api-fails "release-branch $SHA" fail 1 0
run_case odd-count "release-branch $SHA" null 1 0
run_case odd-head "release-branch not-a-sha" 0 1 0

echo "release pull request CI: 1 dispatched, 2 skipped, 3 refused"
