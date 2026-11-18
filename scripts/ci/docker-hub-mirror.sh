#!/usr/bin/env bash

# Routes docker.io pulls through a pull-through mirror (default mirror.gcr.io)
# so hosted runners stop hitting Docker Hub's anonymous pull limit.
# Pinned digests stay valid: the mirror serves the same content addresses.
#
# Usage:
#   docker-hub-mirror.sh                 configure the local Docker daemon
#   docker-hub-mirror.sh --print FILE    print FILE's daemon.json with the
#                                        mirror merged in (no root, no daemon)

set -euo pipefail

mirror="${DOCKER_HUB_MIRROR:-https://mirror.gcr.io}"
readonly mirror

# merge_config reads daemon.json on stdin and keeps every existing key, adding
# the mirror to registry-mirrors once.
merge_config() {
	jq --arg m "$mirror" '.["registry-mirrors"] = (((.["registry-mirrors"] // []) + [$m]) | unique)'
}

if [[ "${1:-}" == --print ]]; then
	if [[ $# -ne 2 ]]; then
		printf 'usage: %s [--print FILE]\n' "$0" >&2
		exit 2
	fi
	if [[ -s "$2" ]]; then
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

if [[ "$(uname -s)" != Linux ]]; then
	printf 'docker-hub-mirror: skip: %s runners run no Linux Docker daemon to configure\n' "$(uname -s)"
	exit 0
fi
if ! command -v docker >/dev/null 2>&1; then
	printf 'docker-hub-mirror: skip: docker is not installed on this runner\n'
	exit 0
fi

conf=/etc/docker/daemon.json
current='{}'
if sudo test -s "$conf"; then
	current="$(sudo cat "$conf")"
fi
merged="$(printf '%s' "$current" | merge_config)"
printf '%s\n' "$merged" | sudo tee "$conf" >/dev/null
sudo systemctl restart docker

for _ in $(seq 1 30); do
	if docker info >/dev/null 2>&1; then
		break
	fi
	sleep 1
done
mirrors="$(docker info --format '{{json .RegistryConfig.Mirrors}}')"
if [[ "$mirrors" != *"${mirror#https://}"* ]]; then
	printf '::error title=Docker Hub mirror::the daemon did not load %s (mirrors: %s)\n' "$mirror" "$mirrors"
	exit 1
fi
printf 'docker-hub-mirror: docker.io pulls go through %s\n' "$mirror"
