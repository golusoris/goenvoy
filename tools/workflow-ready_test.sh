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

# Why: case m: indentless pull_request must be refused.
mkdir -p "$WORK/case_m/.github/workflows"
printf 'name: CI\non:\n- pull_request\n' > "$WORK/case_m/.github/workflows/ci.yml"
expect_refusal "case_m: indentless pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_m"

# Why: case n: indentless push and workflow_dispatch must pass.
mkdir -p "$WORK/case_n/.github/workflows"
printf 'name: CI\non:\n- push\n- workflow_dispatch\n' > "$WORK/case_n/.github/workflows/ci.yml"
expect_pass "case_n: indentless push and workflow_dispatch" "$WORK/case_n"

# Why: case o: CRLF line endings with indentless pull_request must be refused.
mkdir -p "$WORK/case_o/.github/workflows"
printf 'name: CI\r\non:\r\n- pull_request\r\n' > "$WORK/case_o/.github/workflows/ci.yml"
expect_refusal "case_o: CRLF indentless pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_o"

# Why: case p: indentless bare-dash multiline item must be refused.
mkdir -p "$WORK/case_p/.github/workflows"
printf 'name: CI\non:\n-\n  pull_request\n' > "$WORK/case_p/.github/workflows/ci.yml"
expect_refusal "case_p: indentless bare-dash multiline item" "pull_request trigger lacks ready_for_review" "$WORK/case_p"

# Why: case q: indentless on closed at jobs must not treat job ID as pull_request trigger.
mkdir -p "$WORK/case_q/.github/workflows"
printf 'name: CI\non:\n- push\njobs:\n  pull_request:\n    runs-on: ubuntu-latest\n' > "$WORK/case_q/.github/workflows/ci.yml"
expect_pass "case_q: indentless on closed at jobs" "$WORK/case_q"

# Why: case r: top-level key starting with dash must close on and not leak nested pull_request.
mkdir -p "$WORK/case_r/.github/workflows"
printf 'name: CI\non:\n- push\n-custom:\n  pull_request:\n    branches: [main]\n' > "$WORK/case_r/.github/workflows/ci.yml"
expect_pass "case_r: top-level dash key does not keep on open" "$WORK/case_r"

# Why: case s: indentless bare-dash push item must pass.
mkdir -p "$WORK/case_s/.github/workflows"
printf 'name: CI\non:\n-\n  push\n' > "$WORK/case_s/.github/workflows/ci.yml"
expect_pass "case_s: indentless bare-dash push" "$WORK/case_s"

# Why: case t: indentless dash-tab pull_request must be refused.
mkdir -p "$WORK/case_t/.github/workflows"
printf 'name: CI\non:\n-\tpull_request\n' > "$WORK/case_t/.github/workflows/ci.yml"
expect_refusal "case_t: indentless dash-tab pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_t"

# Why: case u: indented dash-tab pull_request must be refused.
mkdir -p "$WORK/case_u/.github/workflows"
printf 'name: CI\non:\n  -\tpull_request\n' > "$WORK/case_u/.github/workflows/ci.yml"
expect_refusal "case_u: indented dash-tab pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_u"

# Why: case v: indented dash-tab ready_for_review in types must pass.
mkdir -p "$WORK/case_v/.github/workflows"
printf 'name: CI\non:\n  pull_request:\n    types:\n      -\tready_for_review\n' > "$WORK/case_v/.github/workflows/ci.yml"
expect_pass "case_v: indented dash-tab ready_for_review in types" "$WORK/case_v"

# Why: case w: indentless dash-tab push must pass.
mkdir -p "$WORK/case_w/.github/workflows"
printf 'name: CI\non:\n-\tpush\n' > "$WORK/case_w/.github/workflows/ci.yml"
expect_pass "case_w: indentless dash-tab push" "$WORK/case_w"

# Why: case x: trailing tab before comment on pull_request item must be refused.
mkdir -p "$WORK/case_x/.github/workflows"
printf 'name: CI\non:\n  - pull_request\t# note\n' > "$WORK/case_x/.github/workflows/ci.yml"
expect_refusal "case_x: trailing tab before comment on pull_request item" "pull_request trigger lacks ready_for_review" "$WORK/case_x"

# Why: case y: trailing tab before comment on multiline bare-dash pull_request item must be refused.
mkdir -p "$WORK/case_y/.github/workflows"
printf 'name: CI\non:\n-\n  pull_request\t# note\n' > "$WORK/case_y/.github/workflows/ci.yml"
expect_refusal "case_y: trailing tab before comment on multiline bare-dash pull_request" "pull_request trigger lacks ready_for_review" "$WORK/case_y"

# Why: case z: trailing tab before comment on ready_for_review in types must pass.
mkdir -p "$WORK/case_z/.github/workflows"
printf 'name: CI\non:\n  pull_request:\n    types:\n      - ready_for_review\t# note\n' > "$WORK/case_z/.github/workflows/ci.yml"
expect_pass "case_z: trailing tab before comment on ready_for_review in types" "$WORK/case_z"

# Why: verify the real repository passes.
REPO_ROOT="$(cd "$HERE/.." && pwd)"
expect_pass "repo: real check" "$REPO_ROOT"
(cd "$REPO_ROOT" && expect_pass "repo: default root")

echo "workflow-ready: all planted cases pass"
