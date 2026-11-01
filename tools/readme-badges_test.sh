#!/usr/bin/env bash
# Planted cases for tools/readme-badges.sh: workflow badges and links
# pointing at existing workflows pass; missing workflows or broken targets
# fail and name the missing workflow file.
#   usage: tools/readme-badges_test.sh
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# expect_refusal NAME NEEDLE ARGS...: the script must fail and name NEEDLE.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$(bash "$HERE/readme-badges.sh" "$@" 2>&1)"; then
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
  out="$(bash "$HERE/readme-badges.sh" "$@")"
  if [ "$out" != "$want" ]; then
    printf 'FAIL %s: got %s, want %s\n' "$name" "$out" "$want" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

# Case (a): badge to existing workflow -> 0
mkdir -p "$WORK/case_a/.github/workflows"
touch "$WORK/case_a/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_a/README.md"
# Project

[![CI](https://github.com/example/repo/actions/workflows/ci.yml/badge.svg)](https://github.com/example/repo/actions/workflows/ci.yml)
EOF
expect_output "case_a: badge to existing workflow" "readme-badges: 2 workflow references ok" "$WORK/case_a"

# Case (b): badge to missing workflow -> 1 and output names the file
mkdir -p "$WORK/case_b/.github/workflows"
cat <<'EOF' > "$WORK/case_b/README.md"
# Project

[![Status](https://github.com/example/repo/actions/workflows/missing.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_b: badge to missing workflow" "missing.yml" "$WORK/case_b"

# Case (c): badge ok but link target missing -> 1
mkdir -p "$WORK/case_c/.github/workflows"
touch "$WORK/case_c/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_c/README.md"
# Project

[![CI](https://github.com/example/repo/actions/workflows/ci.yml/badge.svg)](https://github.com/example/repo/actions/workflows/target-missing.yml)
EOF
expect_refusal "case_c: badge ok but link target missing" "target-missing.yml" "$WORK/case_c"

# Case (d): two md files, miss in second -> 1
mkdir -p "$WORK/case_d/.github/workflows"
touch "$WORK/case_d/.github/workflows/ci.yml"
cat <<'EOF' > "$WORK/case_d/README.md"
# Project

[![CI](https://github.com/example/repo/actions/workflows/ci.yml/badge.svg)](https://github.com/example/repo/actions/workflows/ci.yml)
EOF
cat <<'EOF' > "$WORK/case_d/DOCS.md"
# Docs

[![Audit](https://github.com/example/repo/actions/workflows/audit-missing.yml/badge.svg)](https://example.com)
EOF
expect_refusal "case_d: two md files, miss in second" "audit-missing.yml" "$WORK/case_d"

# Case (e): no workflow refs -> 0
mkdir -p "$WORK/case_e"
cat <<'EOF' > "$WORK/case_e/README.md"
# Project

No workflow badges or links here.
EOF
expect_output "case_e: no workflow refs" "readme-badges: 0 workflow references ok" "$WORK/case_e"

# Real check on the repo root -> 0
REPO_ROOT="$(cd "$HERE/.." && pwd)"
expect_output "repo: real check" "readme-badges: 4 workflow references ok" "$REPO_ROOT"
(cd "$REPO_ROOT" && expect_output "repo: default root" "readme-badges: 4 workflow references ok")

echo "readme-badges: all planted cases pass"
