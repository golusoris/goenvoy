#!/usr/bin/env bash
# Why: shared test assertions for tools/ test scripts.
set -euo pipefail

# expect_refusal NAME NEEDLE CMD...: assert command fails and error message contains needle.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$("$@" 2>&1)"; then
    printf 'FAIL %s: accepted\n%s\n' "$name" "$out" >&2
    exit 1
  fi
  case "$out" in
    *"$needle"*) echo "ok   $name: refused ($needle)" ;;
    *)
      printf 'FAIL %s: refused without "%s"\n%s\n' "$name" "$needle" "$out" >&2
      exit 1
      ;;
  esac
}

# expect_pass NAME CMD...: assert command succeeds.
expect_pass() {
  local name="$1" out
  shift 1
  if ! out="$("$@" 2>&1)"; then
    printf 'FAIL %s: command failed\n%s\n' "$name" "$out" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

# expect_output NAME WANT CMD...: assert command succeeds and stdout equals want.
expect_output() {
  local name="$1" want="$2" out
  shift 2
  if ! out="$("$@")"; then
    printf 'FAIL %s: command failed\n' "$name" >&2
    exit 1
  fi
  if [ "$out" != "$want" ]; then
    printf 'FAIL %s: got %s, want %s\n' "$name" "$out" "$want" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}
