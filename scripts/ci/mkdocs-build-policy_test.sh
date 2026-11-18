#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly repo_root
suite_root="$(mktemp -d "${TMPDIR:-/tmp}/goenvoy-mkdocs-policy.XXXXXX")"
readonly suite_root
trap 'rm -rf "$suite_root"' EXIT

# expect_failure runs a command that must fail and print the expected diagnostic.
expect_failure() {
	local expected="$1"
	shift
	local output
	if output="$("$@" 2>&1)"; then
		printf 'expected failure: %s\n' "$*" >&2
		exit 1
	fi
	if [[ "$output" != *"$expected"* ]]; then
		printf 'missing diagnostic %q in: %s\n' "$expected" "$output" >&2
		exit 1
	fi
}

# new_fixture writes a small mkdocs project of the named shape and prints its root.
new_fixture() {
	local name="$1"
	local shape="$2"
	local root="$suite_root/$name"
	mkdir -p "$root/docs"
	cat >"$root/mkdocs.yml" <<'EOF'
site_name: Policy fixture
theme:
  name: material
nav:
  - Home: index.md
EOF
	case "$shape" in
	negative/clean-page.md)
		printf '# Policy fixture\n\nClean page.\n' >"$root/docs/index.md"
		;;
	positive/broken-link.md)
		printf '# Policy fixture\n\nSee [missing](missing.md).\n' >"$root/docs/index.md"
		;;
	gap/external-link.md)
		printf '# Policy fixture\n\nSee [example](https://example.com).\n' >"$root/docs/index.md"
		;;
	*)
		printf 'unknown fixture shape: %s\n' "$shape" >&2
		exit 2
		;;
	esac
	printf '%s\n' "$root"
}

fake_docker="$suite_root/docker"
cat >"$fake_docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >"${FAKE_DOCKER_LOG:?}"
if [[ "${FAKE_DOCKER_FAIL:-false}" == true ]]; then
	printf 'synthetic Docker failure\n' >&2
	exit 9
fi
site_mount=""
previous=""
for arg in "$@"; do
	if [[ "$previous" == -v && "$arg" == *:/site ]]; then
		site_mount="${arg%:/site}"
		break
	fi
	previous="$arg"
done
[[ -n "$site_mount" ]]
printf '<html></html>\n' >"$site_mount/index.html"
EOF
chmod +x "$fake_docker"

clean_root="$(new_fixture fake-clean negative/clean-page.md)"
fake_log="$suite_root/docker.log"
FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" >/dev/null
for required in \
	'--network none' \
	'--read-only' \
	'--cap-drop ALL' \
	'squidfunk/mkdocs-material:9.7.7@sha256:51b87149d227691486b5f08993d28c65ca7e4990010664b697265b8e6fcd5287' \
	'build --strict --config-file /docs/mkdocs.yml --site-dir /site'; do
	if ! grep -Fq -- "$required" "$fake_log"; then
		printf 'MkDocs Docker contract lacks: %s\n' "$required" >&2
		exit 1
	fi
done

failure_root="$(new_fixture docker-failure negative/clean-page.md)"
expect_failure 'synthetic Docker failure' \
	env FAKE_DOCKER_LOG="$fake_log" FAKE_DOCKER_FAIL=true DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$failure_root"

symlink_root="$(new_fixture symlink negative/clean-page.md)"
ln -s index.md "$symlink_root/docs/alias.md"
expect_failure 'MkDocs build refuses symlink input' \
	env FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$symlink_root"

positive_root="$(new_fixture hiss-positive positive/broken-link.md)"
expect_failure "contains a link 'missing.md'" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$positive_root"
for shape in negative/clean-page.md gap/external-link.md; do
	hiss_root="$(new_fixture "hiss-${shape%%/*}" "$shape")"
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$hiss_root" >/dev/null
done

printf 'MkDocs hermetic, strict, fail-closed policy and HISS fixtures passed\n'
