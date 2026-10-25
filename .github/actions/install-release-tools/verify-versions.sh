#!/usr/bin/env bash
# Why: the installed cosign and Syft must match the pins in action.yml, the one place Renovate bumps.
set -euo pipefail

action="${1:-${GITHUB_ACTION_PATH:?}/action.yml}"

pin() {
  sed -nE "s/^[[:space:]]*$1:[[:space:]]*v?([0-9]+\.[0-9]+\.[0-9]+)[[:space:]]*$/\1/p" "$action"
}

cosign_pin="$(pin cosign-release)"
syft_pin="$(pin syft-version)"
if [ -z "$cosign_pin" ] || [ -z "$syft_pin" ]; then
  echo "no cosign-release or syft-version pin found in $action" >&2
  exit 1
fi

command -v cosign
command -v syft
cosign_got="$(cosign version | sed -nE 's/^GitVersion:[[:space:]]+v?([0-9]+\.[0-9]+\.[0-9]+)(\+dirty)?[[:space:]]*$/\1/p')"
syft_got="$(syft version | sed -nE 's/^Version:[[:space:]]+v?([0-9]+\.[0-9]+\.[0-9]+)[[:space:]]*$/\1/p')"

echo "cosign pin=$cosign_pin installed=$cosign_got"
echo "syft pin=$syft_pin installed=$syft_got"
if [ "$cosign_got" != "$cosign_pin" ] || [ "$syft_got" != "$syft_pin" ]; then
  echo "installed release tools do not match the pins in $action" >&2
  exit 1
fi
