#!/usr/bin/env bash
# Exercises mirror merging and validation without a Docker daemon.
set -euo pipefail
root="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
script="$root/scripts/ci/docker-hub-mirror.sh"

# expect_rejected checks the exit code for an invalid input file or mirror.
expect_rejected() {
	local expected="$1" file="$2" status=0
	shift 2
	"$@" bash "$script" --print "$file" >"$work/out" 2>"$work/err" || status=$?
	if [[ "$status" != "$expected" ]]; then
		printf 'expected exit %s, got %s for %s\n' "$expected" "$status" "$file" >&2
		exit 1
	fi
}
printf '%s\n' '{"registry-mirrors":["https://z", "https://a"]}' >"$work/order"
bash "$script" --print "$work/order" | jq -e '."registry-mirrors" == ["https://z", "https://a", "https://mirror.gcr.io"]' >/dev/null
printf '%s\n' '{"registry-mirrors":["https://z", "https://mirror.gcr.io", "https://a"]}' >"$work/present"
bash "$script" --print "$work/present" | jq -e '."registry-mirrors" == ["https://z", "https://mirror.gcr.io", "https://a"]' >/dev/null
printf '\n' >"$work/newline"
printf '{}\n{}\n' >"$work/two"
printf '[]\n' >"$work/array"
for file in newline two array; do
	expect_rejected 1 "$work/$file" env
done
bash "$script" --print "$work/missing" | jq -e '."registry-mirrors" == ["https://mirror.gcr.io"]' >/dev/null
expect_rejected 2 "$work/missing" env DOCKER_HUB_MIRROR=not-a-url
printf 'Docker Hub mirror print policy passed\n'
