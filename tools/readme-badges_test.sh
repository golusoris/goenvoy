#!/usr/bin/env bash
# Planted cases for tools/readme-badges.sh: workflow badges and links
# pointing at existing workflows pass; missing workflows or broken targets
# fail and name the missing workflow file.
#   usage: tools/readme-badges_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Why: run git without caller repository-location variables (e.g. when run from a git hook).
clean_git() {
  env -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_COMMON_DIR -u GIT_OBJECT_DIRECTORY -u GIT_NAMESPACE git "$@"
}

# expect_refusal NAME NEEDLE ARGS...: the script must fail and name NEEDLE.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  # Why: callers can override the script path to test copies in other directories.
  local script="${SCRIPT_PATH:-$HERE/readme-badges.sh}"
  if out="$(bash "$script" "$@" 2>&1)"; then
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

# expect_output NAME WANT ARGS...: the script must succeed and print WANT.
expect_output() {
  local name="$1" want="$2" out
  shift 2
  if ! out="$(bash "$HERE/readme-badges.sh" "$@")"; then
    printf 'FAIL %s: command failed\n' "$name" >&2
    exit 1
  fi
  if [ "$out" != "$want" ]; then
    printf 'FAIL %s: got "%s", want "%s"\n' "$name" "$out" "$want" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

# expect_pattern NAME PATTERN ARGS...: the script must succeed and match PATTERN.
expect_pattern() {
  local name="$1" pattern="$2" out
  shift 2
  if ! out="$(bash "$HERE/readme-badges.sh" "$@")"; then
    printf 'FAIL %s: command failed\n' "$name" >&2
    exit 1
  fi
  case "$out" in
    *"$pattern"*) echo "ok   $name: $out" ;;
    *)
      printf 'FAIL %s: got "%s", expected pattern "%s"\n' "$name" "$out" "$pattern" >&2
      exit 1
      ;;
  esac
}

# Case (a): badge to existing workflow -> 0
mkdir -p "$WORK/case_a/.github/workflows"
touch "$WORK/case_a/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_a/README.md"
# Project

[![CI](https://github.com/acme/widgets/actions/workflows/ci.yml/badge.svg)](https://github.com/acme/widgets/actions/workflows/ci.yml)
EOF
expect_output "case_a: badge to existing workflow" "readme-badges: 2 workflow references ok" --repo acme/widgets "$WORK/case_a"

# Case (b): badge to missing workflow -> 1 and output names the file
mkdir -p "$WORK/case_b/.github/workflows"
cat <<'EOF' > "$WORK/case_b/README.md"
# Project

[![Status](https://github.com/acme/widgets/actions/workflows/missing.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_b: badge to missing workflow" "missing.yml" --repo acme/widgets "$WORK/case_b"

# Case (c): badge ok but link target missing -> 1
mkdir -p "$WORK/case_c/.github/workflows"
touch "$WORK/case_c/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_c/README.md"
# Project

[![CI](https://github.com/acme/widgets/actions/workflows/ci.yml/badge.svg)](https://github.com/acme/widgets/actions/workflows/target-missing.yml)
EOF
expect_refusal "case_c: badge ok but link target missing" "target-missing.yml" --repo acme/widgets "$WORK/case_c"

# Case (d): two md files, miss in second -> 1
mkdir -p "$WORK/case_d/.github/workflows"
touch "$WORK/case_d/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_d/README.md"
# Project

[![CI](https://github.com/acme/widgets/actions/workflows/ci.yml/badge.svg)](https://github.com/acme/widgets/actions/workflows/ci.yml)
EOF
cat <<'EOF' > "$WORK/case_d/DOCS.md"
# Docs

[![Audit](https://github.com/acme/widgets/actions/workflows/audit-missing.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_d: two md files, miss in second" "audit-missing.yml" --repo acme/widgets "$WORK/case_d"

# Case (e): no workflow refs -> 0
mkdir -p "$WORK/case_e"
cat <<'EOF' > "$WORK/case_e/README.md"
# Project

No workflow badges or links here.
EOF
expect_output "case_e: no workflow refs" "readme-badges: 0 workflow references ok" --repo acme/widgets "$WORK/case_e"

# Case (f): link with /runs suffix to missing workflow -> 1
mkdir -p "$WORK/case_f/.github/workflows"
cat <<'EOF' > "$WORK/case_f/README.md"
# Project

[Runs](https://github.com/acme/widgets/actions/workflows/missing.yml/runs)
EOF
expect_refusal "case_f: link with /runs suffix to missing workflow" "missing.yml" --repo acme/widgets "$WORK/case_f"

# Case (g): link with slash suffix -> 1
mkdir -p "$WORK/case_g/.github/workflows"
cat <<'EOF' > "$WORK/case_g/README.md"
# Project

[Link](https://github.com/acme/widgets/actions/workflows/gone.yml/) more
EOF
expect_refusal "case_g: link with slash suffix" "gone.yml" --repo acme/widgets "$WORK/case_g"

# Case (h): 150 refs on one line, the last missing -> 1
mkdir -p "$WORK/case_h/.github/workflows"
touch "$WORK/case_h/.github/workflows/wf.yml"
{
  printf '# Project\n\n'
  for ((i = 0; i < 149; i++)); do
    printf '[wf](https://github.com/acme/widgets/actions/workflows/wf.yml) '
  done
  printf '[last](https://github.com/acme/widgets/actions/workflows/last-missing.yml)\n'
} > "$WORK/case_h/README.md"
expect_refusal "case_h: 150 refs on one line, last missing" "last-missing.yml" --repo acme/widgets "$WORK/case_h"

# Case (i): DOCS.markdown with missing ref -> 1
mkdir -p "$WORK/case_i/.github/workflows"
cat <<'EOF' > "$WORK/case_i/DOCS.markdown"
# Docs

[CI](https://github.com/acme/widgets/actions/workflows/missing.yml)
EOF
expect_refusal "case_i: DOCS.markdown with missing ref" "missing.yml" --repo acme/widgets "$WORK/case_i"

# Case (j): badge of other repo other/repo missing file -> 0
mkdir -p "$WORK/case_j/.github/workflows"
cat <<'EOF' > "$WORK/case_j/README.md"
# Project

[![External](https://github.com/other/repo/actions/workflows/missing.yml/badge.svg)](https://example.com)
EOF
expect_output "case_j: badge of other repo ignored" "readme-badges: 0 workflow references ok" --repo acme/widgets "$WORK/case_j"

# Case (k): bare URL ending with sentence punctuation -> 0
mkdir -p "$WORK/case_k/.github/workflows"
touch "$WORK/case_k/.github/workflows/ok.yml"
cat <<'EOF' > "$WORK/case_k/README.md"
# Project

see https://github.com/acme/widgets/actions/workflows/ok.yml.
EOF
expect_output "case_k: bare URL with sentence punctuation" "readme-badges: 1 workflow references ok" --repo acme/widgets "$WORK/case_k"

# Case (l): relative badge ../../actions/workflows/gone.yml/badge.svg -> 1
mkdir -p "$WORK/case_l/.github/workflows"
cat <<'EOF' > "$WORK/case_l/README.md"
# Project

[![Badge](../../actions/workflows/gone.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_l: relative badge missing" "gone.yml" --repo acme/widgets "$WORK/case_l"

# Case (m): link to ci.yml-old where only ci.yml exists -> 1
mkdir -p "$WORK/case_m/.github/workflows"
touch "$WORK/case_m/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_m/README.md"
# Project

[CI Old](https://github.com/acme/widgets/actions/workflows/ci.yml-old)
EOF
expect_refusal "case_m: link to ci.yml-old where only ci.yml exists" "ci.yml-old" --repo acme/widgets "$WORK/case_m"

# Case (n): double-quoted img src relative to missing gone.yml -> 1
mkdir -p "$WORK/case_n/.github/workflows"
cat <<'EOF' > "$WORK/case_n/README.md"
# Project

<img src="../../actions/workflows/gone.yml/badge.svg">
EOF
expect_refusal "case_n: double-quoted img src relative to missing gone.yml" "gone.yml" --repo acme/widgets "$WORK/case_n"

# Case (o): single-quoted href full URL acme/widgets missing -> 1
mkdir -p "$WORK/case_o/.github/workflows"
cat <<'EOF' > "$WORK/case_o/README.md"
# Project

<a href='https://github.com/acme/widgets/actions/workflows/missing.yml'>CI</a>
EOF
expect_refusal "case_o: single-quoted href full URL acme/widgets missing" "missing.yml" --repo acme/widgets "$WORK/case_o"

# Case (p): default-slug detection from git origin -> 1
mkdir -p "$WORK/case_p/tools" "$WORK/case_p/.github/workflows"
clean_git -C "$WORK/case_p" init -q
clean_git -C "$WORK/case_p" remote add origin https://github.com/acme/widgets.git
cp "$HERE/readme-badges.sh" "$WORK/case_p/tools/readme-badges.sh"
cat <<'EOF' > "$WORK/case_p/README.md"
# Project

[CI](https://github.com/acme/widgets/actions/workflows/missing.yml)
EOF
if out="$(bash "$WORK/case_p/tools/readme-badges.sh" 2>&1)"; then
  printf 'FAIL %s: accepted\n%s\n' "case_p: default-slug detection" "$out" >&2
  exit 1
fi
case "$out" in
  *"missing.yml"*) echo "ok   case_p: default-slug detection: refused (missing.yml)" ;;
  *)
    printf 'FAIL %s: refused without "%s"\n%s\n' "case_p: default-slug detection" "missing.yml" "$out" >&2
    exit 1
    ;;
esac

# Case (q): scanned root slug detected from root arg git origin -> 1
# Why: without --repo, slug must come from scanned root git origin, not script checkout.
mkdir -p "$WORK/case_q_script/tools" "$WORK/case_q/.github/workflows"
cp "$HERE/readme-badges.sh" "$WORK/case_q_script/tools/readme-badges.sh"
clean_git -C "$WORK/case_q" init -q
clean_git -C "$WORK/case_q" remote add origin https://github.com/acme/widgets.git
cat <<'EOF' > "$WORK/case_q/README.md"
# Project

[CI](https://github.com/acme/widgets/actions/workflows/missing.yml)
EOF
SCRIPT_PATH="$WORK/case_q_script/tools/readme-badges.sh" \
  expect_refusal "case_q: scanned root slug from target root origin" "missing.yml" "$WORK/case_q"

# Case (r): scanned root slug detected when caller exports GIT_DIR to a different repository -> 1
# Why: git hooks export GIT_DIR; slug must come from scanned root git origin, not caller GIT_DIR.
mkdir -p "$WORK/case_r_script/tools" "$WORK/case_r/.github/workflows" "$WORK/case_r_other"
cp "$HERE/readme-badges.sh" "$WORK/case_r_script/tools/readme-badges.sh"
clean_git -C "$WORK/case_r" init -q
clean_git -C "$WORK/case_r" remote add origin https://github.com/acme/widgets.git
clean_git -C "$WORK/case_r_other" init -q
clean_git -C "$WORK/case_r_other" remote add origin https://github.com/other/repo.git
cat <<'EOF' > "$WORK/case_r/README.md"
# Project

[CI](https://github.com/acme/widgets/actions/workflows/missing.yml)
EOF
GIT_DIR="$WORK/case_r_other/.git" \
  SCRIPT_PATH="$WORK/case_r_script/tools/readme-badges.sh" \
  expect_refusal "case_r: scanned root slug with foreign GIT_DIR" "missing.yml" "$WORK/case_r"

# Real check on the repo root -> 0
REPO_ROOT="$(cd "$HERE/.." && pwd)"
expect_pattern "repo: real check" "workflow references ok" --repo golusoris/goenvoy "$REPO_ROOT"
(cd "$REPO_ROOT" && expect_pattern "repo: default root" "workflow references ok" --repo golusoris/goenvoy)

echo "readme-badges: all planted cases pass"
