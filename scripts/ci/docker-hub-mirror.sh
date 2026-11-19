#!/usr/bin/env bash

# Routes docker.io pulls through a pull-through mirror without changing digest pins.
# Usage: docker-hub-mirror.sh [--print FILE]
set -euo pipefail

mirror="${DOCKER_HUB_MIRROR:-https://mirror.gcr.io}"
readonly mirror
if [[ ! "$mirror" =~ ^https?://[^/[:space:]]+/?$ ]]; then
	printf 'docker-hub-mirror: invalid mirror URL: %s\n' "$mirror" >&2
	exit 2
fi

# merge_config appends the mirror while preserving the existing order.
merge_config() {
	jq --arg m "$mirror" '.["registry-mirrors"] = ((.["registry-mirrors"] // []) as $a | if ($a | index($m)) == null then $a + [$m] else $a end)'
}

# validate_config requires exactly one JSON object.
validate_config() {
	jq -se 'length == 1 and (.[0] | type == "object")' "$1" >/dev/null
}

# install_config stages the configuration before atomically replacing the target.
install_config() {
	local source="$1" target="$2" had_previous="$3" privilege="${DOCKER_MIRROR_PRIVILEGE:-sudo}" step=""
	if ! "$privilege" install -m 0644 "$source" "$target.new"; then
		step=install
	elif [[ "$had_previous" == true ]] && ! "$privilege" chown --reference="$target" "$target.new"; then
		step='chown'
	elif [[ "$had_previous" == true ]] && ! "$privilege" chmod --reference="$target" "$target.new"; then
		step='chmod'
	elif ! "$privilege" mv -f "$target.new" "$target"; then
		step=rename
	fi
	if [[ -n "$step" ]]; then
		if ! "$privilege" rm -f "$target.new"; then
			step="$step; staged file cleanup failed"
		fi
		printf '::error title=Docker Hub mirror::%s failed\n' "$step" >&2
		return 1
	fi
}

if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
	return
fi

if [[ "${1:-}" == --print ]]; then
	if [[ $# -ne 2 ]]; then
		printf 'usage: %s [--print FILE]\n' "$0" >&2
		exit 2
	fi
	if [[ -s "$2" ]]; then
		if ! validate_config "$2"; then
			printf 'docker-hub-mirror: expected exactly one JSON object: %s\n' "$2" >&2
			exit 1
		fi
		merge_config <"$2"
	else
		printf '{}' | merge_config
	fi
	exit 0
fi
if [[ $# -ne 0 ]]; then
	printf 'usage: %s [--print FILE]\n' "$0" >&2
	exit 2
fi
if [[ "$(uname -s)" != Linux ]] || ! command -v docker >/dev/null 2>&1; then
	printf 'docker-hub-mirror: skip: Linux Docker daemon unavailable\n'
	exit 0
fi

conf=/etc/docker/daemon.json
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
previous=false
if sudo test -e "$conf"; then
	previous=true
	sudo cat "$conf" | tee "$work/previous" >/dev/null
fi
if [[ "$previous" == true && -s "$work/previous" ]]; then
	if ! validate_config "$work/previous"; then
		printf '::error title=Docker Hub mirror::existing configuration is not one JSON object\n' >&2
		exit 1
	fi
	merge_config <"$work/previous" >"$work/merged"
else
	printf '{}' | merge_config >"$work/merged"
fi
if ! validate_config "$work/merged"; then
	printf '::error title=Docker Hub mirror::merged configuration is not one JSON object\n' >&2
	exit 1
fi

# restore_config restores the original file and restarts Docker after failure.
restore_config() {
	local outcome='configuration restored'
	if [[ "$previous" == true ]]; then
		if ! install_config "$work/previous" "$conf" true; then
			outcome='restoring the previous configuration failed'
		fi
	elif ! sudo rm -f "$conf"; then
		outcome='removing the new configuration failed'
	fi
	if ! sudo systemctl restart docker; then
		outcome="$outcome; Docker did not restart afterwards"
	fi
	printf '::error title=Docker Hub mirror::Docker restart or readiness failed; %s\n' "$outcome" >&2
	exit 1
}

install_config "$work/merged" "$conf" "$previous"
if ! sudo systemctl restart docker; then
	restore_config
fi
ready=false
for ((attempt = 0; attempt < 30; attempt++)); do
	if docker info >/dev/null 2>&1; then
		ready=true
		break
	fi
	sleep 1
done
if [[ "$ready" != true ]]; then
	restore_config
fi
mirrors="$(docker info --format '{{json .RegistryConfig.Mirrors}}')"
if ! jq -e --arg m "${mirror%/}" 'type == "array" and any(.[]; type == "string" and (rtrimstr("/") == $m))' <<<"$mirrors" >/dev/null; then
	printf '::error title=Docker Hub mirror::the daemon did not load %s (mirrors: %s)\n' "$mirror" "$mirrors" >&2
	exit 1
fi
printf 'docker-hub-mirror: docker.io pulls go through %s\n' "$mirror"
