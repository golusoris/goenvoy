#!/usr/bin/env bash
# Exercise privileged write checks against tracked fixture files.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=/dev/null
. "$HERE/testlib.sh"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/.github/workflows"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" init -q
printf "clean\n" > "$WORK/a.sh"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" add a.sh
SUT=(bash "$HERE/privileged-write.sh" "$WORK")
printf "%s\n" 'sudo'' cp a b' > "$WORK/a.sh"
expect_refusal copy "a.sh:1:" "${SUT[@]}"
printf '%s\044%s\n' '"' 'privilege" cp a b' > "$WORK/a.sh"
expect_refusal variable "a.sh:1:" "${SUT[@]}"
printf "%s\n" 'sudo'' install a b > log' > "$WORK/a.sh"
expect_refusal redirect "a.sh:1:" "${SUT[@]}"
printf "%s\n" 'sudo'' tee /etc/x' > "$WORK/a.sh"
expect_refusal tee "a.sh:1:" "${SUT[@]}"
printf "clean\n" > "$WORK/a.sh"
printf "%s\n" 'run: sudo'' cp a b' > "$WORK/.github/workflows/w.yml"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" add .github/workflows/w.yml
expect_refusal workflow ".github/workflows/w.yml:1:" "${SUT[@]}"
printf "clean\n" > "$WORK/.github/workflows/w.yml"
printf "\t%s\n" 'sudo'' cp a b' > "$WORK/Makefile"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" add Makefile
expect_refusal makefile "Makefile:1:" "${SUT[@]}"
printf "clean\n" > "$WORK/Makefile"
printf "%s\n" 'sudo'' install -m 0644 a b.new' 'sudo'' mv -f b.new b' > "$WORK/a.sh"
expect_pass staged "${SUT[@]}"
printf "%s\n" '# sudo'' cp a b' > "$WORK/a.sh"
expect_pass comment "${SUT[@]}"
printf "%s\n" 'sudo'' cp a b' > "$WORK/b.sh"
expect_pass untracked "${SUT[@]}"
printf "%s\n" 'sudo'' cp a b' > "$WORK/README.md"
GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -C "$WORK" add README.md
expect_pass ignored_type "${SUT[@]}"
expect_pass repository bash "$HERE/privileged-write.sh"
echo "privileged-write: all planted cases pass"
