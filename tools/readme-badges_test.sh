#!/usr/bin/env bash
# Planted cases for tools/readme-badges.sh: workflow badges and links
# pointing at existing workflows pass; missing workflows or broken targets
# fail and name the missing workflow file.
#   usage: tools/readme-badges_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=tools/testlib.sh
. "$HERE/testlib.sh"
# Why: callers can override the script path to test copies in other directories.
SUT=(bash "${SCRIPT_PATH:-$HERE/readme-badges.sh}")
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# expect_status NAME STATUS NEEDLE ARGS...: the script must exit STATUS and name NEEDLE.
expect_status() {
  local name="$1" want_status="$2" needle="$3" out status=0
  shift 3
  if out="$("${SUT[@]}" "$@" 2>&1)"; then
    status=0
  else
    status=$?
  fi
  if [ "$status" -ne "$want_status" ]; then
    printf 'FAIL %s: got status %d, want %d\n%s\n' "$name" "$status" "$want_status" "$out" >&2
    exit 1
  fi
  if [ -n "$needle" ]; then
    case "$out" in
      *"$needle"*) echo "ok   $name: refused ($needle, status $status)" ;;
      *)
        printf 'FAIL %s: refused without "%s"\n%s\n' "$name" "$needle" "$out" >&2
        exit 1
        ;;
    esac
  else
    echo "ok   $name: refused (status $status)"
  fi
}

# expect_pattern NAME PATTERN ARGS...: the script must succeed and match PATTERN.
expect_pattern() {
  local name="$1" pattern="$2" out
  shift 2
  if ! out="$("${SUT[@]}" "$@")"; then
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
expect_output "case_a: badge to existing workflow" "readme-badges: 2 workflow references ok" "${SUT[@]}" --repo acme/widgets "$WORK/case_a"

# Case (b): badge to missing workflow -> 1 and output names the file
mkdir -p "$WORK/case_b/.github/workflows"
cat <<'EOF' > "$WORK/case_b/README.md"
# Project

[![Status](https://github.com/acme/widgets/actions/workflows/missing.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_b: badge to missing workflow" "missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_b"

# Case (c): badge ok but link target missing -> 1
mkdir -p "$WORK/case_c/.github/workflows"
touch "$WORK/case_c/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_c/README.md"
# Project

[![CI](https://github.com/acme/widgets/actions/workflows/ci.yml/badge.svg)](https://github.com/acme/widgets/actions/workflows/target-missing.yml)
EOF
expect_refusal "case_c: badge ok but link target missing" "target-missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_c"

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
expect_refusal "case_d: two md files, miss in second" "audit-missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_d"

# Case (e): no workflow refs -> 0
mkdir -p "$WORK/case_e"
cat <<'EOF' > "$WORK/case_e/README.md"
# Project

No workflow badges or links here.
EOF
expect_output "case_e: no workflow refs" "readme-badges: 0 workflow references ok" "${SUT[@]}" --repo acme/widgets "$WORK/case_e"

# Case (f): link with /runs suffix to missing workflow -> 1
mkdir -p "$WORK/case_f/.github/workflows"
cat <<'EOF' > "$WORK/case_f/README.md"
# Project

[Runs](https://github.com/acme/widgets/actions/workflows/missing.yml/runs)
EOF
expect_refusal "case_f: link with /runs suffix to missing workflow" "missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_f"

# Case (g): link with slash suffix -> 1
mkdir -p "$WORK/case_g/.github/workflows"
cat <<'EOF' > "$WORK/case_g/README.md"
# Project

[Link](https://github.com/acme/widgets/actions/workflows/gone.yml/) more
EOF
expect_refusal "case_g: link with slash suffix" "gone.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_g"

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
expect_refusal "case_h: 150 refs on one line, last missing" "last-missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_h"

# Case (i): DOCS.markdown with missing ref -> 1
mkdir -p "$WORK/case_i/.github/workflows"
cat <<'EOF' > "$WORK/case_i/DOCS.markdown"
# Docs

[CI](https://github.com/acme/widgets/actions/workflows/missing.yml)
EOF
expect_refusal "case_i: DOCS.markdown with missing ref" "missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_i"

# Case (j): badge of other repo other/repo missing file -> 0
mkdir -p "$WORK/case_j/.github/workflows"
cat <<'EOF' > "$WORK/case_j/README.md"
# Project

[![External](https://github.com/other/repo/actions/workflows/missing.yml/badge.svg)](https://example.com)
EOF
expect_output "case_j: badge of other repo ignored" "readme-badges: 0 workflow references ok" "${SUT[@]}" --repo acme/widgets "$WORK/case_j"

# Case (k): bare URL ending with sentence punctuation -> 0
mkdir -p "$WORK/case_k/.github/workflows"
touch "$WORK/case_k/.github/workflows/ok.yml"
cat <<'EOF' > "$WORK/case_k/README.md"
# Project

see https://github.com/acme/widgets/actions/workflows/ok.yml.
EOF
expect_output "case_k: bare URL with sentence punctuation" "readme-badges: 1 workflow references ok" "${SUT[@]}" --repo acme/widgets "$WORK/case_k"

# Case (l): relative badge ../../actions/workflows/gone.yml/badge.svg -> 1
mkdir -p "$WORK/case_l/.github/workflows"
cat <<'EOF' > "$WORK/case_l/README.md"
# Project

[![Badge](../../actions/workflows/gone.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_l: relative badge missing" "gone.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_l"

# Case (m): link to ci.yml-old where only ci.yml exists -> 1
mkdir -p "$WORK/case_m/.github/workflows"
touch "$WORK/case_m/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_m/README.md"
# Project

[CI Old](https://github.com/acme/widgets/actions/workflows/ci.yml-old)
EOF
expect_refusal "case_m: link to ci.yml-old where only ci.yml exists" "ci.yml-old" "${SUT[@]}" --repo acme/widgets "$WORK/case_m"

# Case (n): double-quoted img src relative to missing gone.yml -> 1
mkdir -p "$WORK/case_n/.github/workflows"
cat <<'EOF' > "$WORK/case_n/README.md"
# Project

<img src="../../actions/workflows/gone.yml/badge.svg">
EOF
expect_refusal "case_n: double-quoted img src relative to missing gone.yml" "gone.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_n"

# Case (o): single-quoted href full URL acme/widgets missing -> 1
mkdir -p "$WORK/case_o/.github/workflows"
cat <<'EOF' > "$WORK/case_o/README.md"
# Project

<a href='https://github.com/acme/widgets/actions/workflows/missing.yml'>CI</a>
EOF
expect_refusal "case_o: single-quoted href full URL acme/widgets missing" "missing.yml" "${SUT[@]}" --repo acme/widgets "$WORK/case_o"

# Case (u): no --repo -> exit status exactly 2 and output names --repo
mkdir -p "$WORK/case_u"
touch "$WORK/case_u/README.md"
expect_status "case_u: no --repo" 2 "--repo" "$WORK/case_u"

# Case (v): --repo with no slash -> exit status exactly 2
mkdir -p "$WORK/case_v"
touch "$WORK/case_v/README.md"
expect_status "case_v: --repo with no slash" 2 "" --repo widgets "$WORK/case_v"

# Real check on the repo root -> 0
REPO_ROOT="$(cd "$HERE/.." && pwd)"
expect_pattern "repo: real check" "workflow references ok" --repo golusoris/goenvoy "$REPO_ROOT"
(cd "$REPO_ROOT" && expect_pattern "repo: default root" "workflow references ok" --repo golusoris/goenvoy)

echo "readme-badges: all planted cases pass"
