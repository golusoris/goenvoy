#!/usr/bin/env bash
# Exercises mirror merging and validation without a Docker daemon.
set -euo pipefail
root="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
script="$root/scripts/ci/docker-hub-mirror.sh"

# check_no_inplace_privileged_write rejects unapproved privileged commands and redirects.
check_no_inplace_privileged_write() {
	awk '
		BEGIN {
			prefix = "(^|[;&|([:space:]])(" "su" "do" "|\"\\$" "privilege\")"
			command = prefix "[[:space:]]+[^[:space:];|&]+"
		}
		FNR > 100000 { print FILENAME ":" FNR; failed = 1; nextfile }
		/^[[:space:]]*#/ { next }
		{
			line = $0
			while (match(line, command)) {
				call = substr(line, RSTART, RLENGTH)
				sub(prefix "[[:space:]]+", "", call)
				rest = substr(line, RSTART + RLENGTH)
				split(rest, segments, /[;|&]/)
				if (call !~ /^(install|mv|rm|chown|chmod|test|cat|systemctl)$/ || segments[1] ~ />/) {
					print FILENAME ":" FNR
					failed = 1
				}
				line = rest
			}
		}
		END { exit failed }
	' "$@"
}

printf '%s %s\n' 'sudo' 'cp a b' >"$work/unsafe"
if check_no_inplace_privileged_write "$work/unsafe" >"$work/scan"; then
	printf 'expected privileged copy rejection\n' >&2
	exit 1
fi
grep -Fxq "$work/unsafe:1" "$work/scan"
printf '%s %s\n' 'sudo' 'install a b.new' 'sudo' 'mv b.new b' >"$work/staged"
check_no_inplace_privileged_write "$work/staged"
check_no_inplace_privileged_write "$root"/scripts/ci/*.sh

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
env DOCKER_HUB_MIRROR=https://mirror.gcr.io bash "$script" --print "$work/order" | jq -e '."registry-mirrors" == ["https://z", "https://a", "https://mirror.gcr.io"]' >/dev/null
printf '%s\n' '{"registry-mirrors":["https://z", "https://mirror.gcr.io", "https://a"]}' >"$work/present"
env DOCKER_HUB_MIRROR=https://mirror.gcr.io bash "$script" --print "$work/present" | jq -e '."registry-mirrors" == ["https://z", "https://mirror.gcr.io", "https://a"]' >/dev/null
printf '\n' >"$work/newline"
printf '{}\n{}\n' >"$work/two"
printf '[]\n' >"$work/array"
for file in newline two array; do
	expect_rejected 1 "$work/$file" env DOCKER_HUB_MIRROR=https://mirror.gcr.io
done
env DOCKER_HUB_MIRROR=https://mirror.gcr.io bash "$script" --print "$work/missing" | jq -e '."registry-mirrors" == ["https://mirror.gcr.io"]' >/dev/null
expect_rejected 2 "$work/missing" env DOCKER_HUB_MIRROR=not-a-url

export DOCKER_HUB_MIRROR=https://mirror.gcr.io DOCKER_MIRROR_PRIVILEGE=env
# shellcheck source=/dev/null
. "$script"
printf 'old\n' >"$work/config"
printf 'new\n' >"$work/merged"
install_config "$work/merged" "$work/config" true
cmp "$work/merged" "$work/config"
[[ ! -e "$work/config.new" ]]
printf 'old\n' >"$work/config"
cp "$work/config" "$work/original"
cat >"$work/fail-install" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == install ]]; then
	printf 'partial\n' >"${@: -1}"
	exit 1
fi
"$@"
EOF
chmod +x "$work/fail-install"
export DOCKER_MIRROR_PRIVILEGE="$work/fail-install"
if install_config "$work/merged" "$work/config" true 2>"$work/err"; then
	printf 'expected staging failure\n' >&2
	exit 1
fi
cmp "$work/original" "$work/config"
[[ ! -e "$work/config.new" ]]
grep -Fq '::error title=Docker Hub mirror::install failed' "$work/err"
printf 'Docker Hub mirror print policy passed\n'
