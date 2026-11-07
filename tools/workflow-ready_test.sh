#!/usr/bin/env bash
# Why: planted test cases for tools/workflow-ready.sh.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Why: assert command fails and error message contains needle.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$(bash "$HERE/workflow-ready.sh" "$@" 2>&1)"; then
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

# Why: assert command succeeds.
expect_pass() {
  local name="$1" out
  shift 1
  if ! out="$(bash "$HERE/workflow-ready.sh" "$@" 2>&1)"; then
    printf 'FAIL %s: command failed\n%s\n' "$name" "$out" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

# Why: case a: pull_request trigger without types list must be refused.
mkdir -p "$WORK/case_a/.github/workflows"
cat <<'EOF' > "$WORK/case_a/.github/workflows/ci.yml"
name: CI
on:
  pull_request:
    branches: [main]
EOF
expect_refusal "case_a: trigger without types" "pull_request trigger lacks ready_for_review" "$WORK/case_a"

# Why: case b: pull_request trigger with types missing ready_for_review must be refused.
mkdir -p "$WORK/case_b/.github/workflows"
cat <<'EOF' > "$WORK/case_b/.github/workflows/ci.yml"
name: CI
on:
  pull_request:
    types: [opened, synchronize]
EOF
expect_refusal "case_b: types without ready_for_review" "pull_request trigger lacks ready_for_review" "$WORK/case_b"

# Why: case c: flow-form types containing ready_for_review must pass.
mkdir -p "$WORK/case_c/.github/workflows"
cat <<'EOF' > "$WORK/case_c/.github/workflows/ci.yml"
name: CI
on:
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
EOF
expect_pass "case_c: flow-form types with ready_for_review" "$WORK/case_c"

# Why: case d: block-list types containing ready_for_review must pass.
mkdir -p "$WORK/case_d/.github/workflows"
cat <<'EOF' > "$WORK/case_d/.github/workflows/ci.yml"
name: CI
on:
  pull_request:
    types:
      - opened
      - synchronize
      - ready_for_review
EOF
expect_pass "case_d: block-list types with ready_for_review" "$WORK/case_d"

# Why: case e: workflow without pull_request trigger must pass.
mkdir -p "$WORK/case_e/.github/workflows"
cat <<'EOF' > "$WORK/case_e/.github/workflows/push.yml"
name: Push
on:
  push:
    branches: [main]
EOF
expect_pass "case_e: no pull_request trigger" "$WORK/case_e"

# Why: case f: only a .yaml workflow with pull_request and no types must be refused.
mkdir -p "$WORK/case_f/.github/workflows"
cat <<'EOF' > "$WORK/case_f/.github/workflows/push.yml"
name: Push
on:
  push:
    branches: [main]
EOF
cat <<'EOF' > "$WORK/case_f/.github/workflows/ci.yaml"
name: CI
on:
  pull_request:
    branches: [main]
EOF
expect_refusal "case_f: only a .yaml workflow with pull_request and no types" ".github/workflows/ci.yaml" "$WORK/case_f"

# Why: case g: .yaml workflow with ready_for_review must pass.
mkdir -p "$WORK/case_g/.github/workflows"
cat <<'EOF' > "$WORK/case_g/.github/workflows/ci.yaml"
name: CI
on:
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
EOF
expect_pass "case_g: .yaml workflow with ready_for_review" "$WORK/case_g"

# Why: case h: block-list event form (- pull_request) must be refused.
mkdir -p "$WORK/case_h/.github/workflows"
cat <<'EOF' > "$WORK/case_h/.github/workflows/ci.yml"
name: CI
on:
  - push
  - pull_request
EOF
expect_refusal "case_h: block-list event pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_h"

# Why: case i: nested list containing pull_request (e.g. branch name) must not
# cause a false positive.
mkdir -p "$WORK/case_i/.github/workflows"
cat <<'EOF' > "$WORK/case_i/.github/workflows/ci.yml"
name: CI
on:
  push:
    branches:
      - pull_request
      - main
EOF
expect_pass "case_i: nested pull_request branch name" "$WORK/case_i"

# Why: case j: CRLF line endings with block-list pull_request must be refused.
mkdir -p "$WORK/case_j/.github/workflows"
printf 'name: CI\r\non:\r\n  - push\r\n  - pull_request\r\n' > "$WORK/case_j/.github/workflows/ci.yml"
expect_refusal "case_j: CRLF block-list event" "pull_request trigger lacks ready_for_review" "$WORK/case_j"

# Why: case k: bare-dash multiline block-sequence item must be refused.
mkdir -p "$WORK/case_k/.github/workflows"
printf 'name: CI\non:\n  -\n    pull_request\n' > "$WORK/case_k/.github/workflows/ci.yml"
expect_refusal "case_k: bare-dash multiline block-sequence" "pull_request trigger lacks ready_for_review" "$WORK/case_k"

# Why: case l: folded scalar block-sequence item must be refused.
mkdir -p "$WORK/case_l/.github/workflows"
printf 'name: CI\non:\n  - >-\n    pull_request\n' > "$WORK/case_l/.github/workflows/ci.yml"
expect_refusal "case_l: folded scalar block-sequence" "pull_request trigger lacks ready_for_review" "$WORK/case_l"

# Why: verify the real repository passes.
REPO_ROOT="$(cd "$HERE/.." && pwd)"
expect_pass "repo: real check" "$REPO_ROOT"
(cd "$REPO_ROOT" && expect_pass "repo: default root")

echo "workflow-ready: all planted cases pass"
