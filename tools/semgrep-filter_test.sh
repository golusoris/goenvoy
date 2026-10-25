#!/usr/bin/env bash
# Planted cases for tools/semgrep-filter.sh: the excepted file and rule pass before
# expiry; another file, another rule, an expired date and an unsuppressed extra fail.
#   usage: tools/semgrep-filter_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

RULE=python.lang.security.audit.dynamic-urllib-use-detected.dynamic-urllib-use-detected
FILE=tools/figures/mkdocs_hook.py

# plant NAME RESULT...: one SARIF run holding the given result objects.
plant() {
  local name=$1
  shift
  local IFS=,
  printf '{"runs":[{"results":[%s]}]}\n' "$*" > "$WORK/$name.sarif"
}

result() {
  printf '{"ruleId":"%s","locations":[{"physicalLocation":{"artifactLocation":{"uri":"%s"},"region":{"startLine":1}}}]%s}' "$1" "$2" "${3:-}"
}

fail=0
expect() {
  local want=$1 name=$2 today=$3 got=0
  bash "$HERE/semgrep-filter.sh" "$WORK/$name.sarif" "$today" 2>/dev/null || got=$?
  if { [ "$want" = pass ] && [ "$got" -ne 0 ]; } || { [ "$want" = fail ] && [ "$got" -eq 0 ]; }; then
    echo "FAIL: $name on $today: want $want, rc $got" >&2
    fail=1
  fi
}

plant empty
expect pass empty 2026-10-25

plant excepted "$(result "$RULE" "$FILE")"
expect pass excepted 2026-10-28

plant excepted-dot "$(result "$RULE" "./$FILE")"
expect pass excepted-dot 2026-10-25

plant expired "$(result "$RULE" "$FILE")"
expect fail expired 2026-10-29

plant other-file "$(result "$RULE" tools/figures/other.py)"
expect fail other-file 2026-10-25

plant other-rule "$(result python.lang.security.audit.other.other "$FILE")"
expect fail other-rule 2026-10-25

plant suppressed "$(result go.lang.x.y main.go ',"suppressions":[{"kind":"inSource"}]')"
expect pass suppressed 2026-10-25

plant mixed "$(result "$RULE" "$FILE")" "$(result go.lang.x.y main.go)"
expect fail mixed 2026-10-25

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "semgrep-filter: planted cases pass"
