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
if [[ "${1:-}" == info ]]; then
	printf '%s\n' "${FAKE_SECURITY_OPTIONS:-[]}" 
	exit 0
fi
printf '%s\n' "$@" >"${FAKE_DOCKER_LOG:?}"
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
if [[ -n "${FAKE_APPEAR_TARGET:-}" ]]; then
	mkdir -p -- "$FAKE_APPEAR_TARGET"
fi
if [[ -n "${FAKE_REPLACE_TARGET:-}" ]]; then
	mv -- "$FAKE_REPLACE_TARGET" "${FAKE_REPLACE_TARGET}.original"
	mkdir -- "$FAKE_REPLACE_TARGET"
	printf 'replacement marker\n' >"$FAKE_REPLACE_TARGET/marker"
fi
if [[ -n "${FAKE_REMOVE_TARGET:-}" ]]; then
	rm -rf -- "$FAKE_REMOVE_TARGET"
fi
if [[ "${FAKE_PARTIAL_FAIL:-false}" == true ]]; then
	printf 'partial\n' >"$site_mount/asset.css"
	exit 1
fi
printf '%s\n' "${FAKE_INDEX:-<html></html>}" >"$site_mount/index.html"
EOF
chmod +x "$fake_docker"

clean_root="$(new_fixture fake-clean negative/clean-page.md)"
fake_log="$suite_root/docker.log"
FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" >/dev/null
# shellcheck source=/dev/null
. "$repo_root/tools/tool-versions.env"
if [[ ! "$MKDOCS_MATERIAL_IMAGE_DIGEST" =~ ^sha256:[[:xdigit:]]{64}$ ]]; then
	printf 'invalid MkDocs image digest\n' >&2
	exit 1
fi
for required in --read-only --strict "squidfunk/mkdocs-material:${MKDOCS_MATERIAL_VERSION}@${MKDOCS_MATERIAL_IMAGE_DIGEST}" -u; do
	if ! grep -Fxq -- "$required" "$fake_log"; then
		printf 'MkDocs Docker contract lacks token: %s\n' "$required" >&2
		exit 1
	fi
done
if ! awk 'prev == "--network" && $0 == "none" { found = 1 } { prev = $0 } END { exit !found }' "$fake_log"; then
	printf 'MkDocs Docker contract lacks --network none\n' >&2
	exit 1
fi

(cd "$suite_root" && CDPATH=. FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root fake-clean >/dev/null)
printf 'docs_dir: manual\n' >>"$clean_root/mkdocs.yml"
expect_failure 'MkDocs refuses docs_dir: manual' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root"
sed -i '/^docs_dir: manual$/d' "$clean_root/mkdocs.yml"
printf '"docs_dir": manual  \n' >>"$clean_root/mkdocs.yml"
expect_failure 'MkDocs refuses docs_dir: manual' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root"
sed -i '/^"docs_dir": manual  $/d' "$clean_root/mkdocs.yml"
printf 'docs_dir: docs  # default\n' >>"$clean_root/mkdocs.yml"
env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" >/dev/null
sed -i '/^docs_dir: docs  # default$/d' "$clean_root/mkdocs.yml"
ln -s docs "$clean_root/site"
expect_failure 'MkDocs output refuses symlink' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir site/.
expect_failure 'MkDocs output overlaps source' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir .
expect_failure 'MkDocs output overlaps source' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir docs/out
ln -s "$clean_root" "$suite_root/linked-root"
expect_failure 'MkDocs output overlaps source' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$suite_root/linked-root" --site-dir "$clean_root/docs/out"
mkdir "$clean_root/foreign"
printf 'foreign\n' >"$clean_root/foreign/file"
expect_failure 'MkDocs output refuses nonempty directory' env DOCKER_BIN="$fake_docker" FAKE_DOCKER_LOG="$fake_log" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir foreign
target="$clean_root/published"
target_lock="$clean_root/.published.lock"
mkdir "$target_lock"
expect_failure "MkDocs output is locked by another build: $target_lock" \
	env FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target"
if [[ -e "$target" || ! -d "$target_lock" || \
	-n "$(find "$clean_root" -maxdepth 1 -name '.published.*' ! -name '.published.lock' -print -quit)" ]]; then
	printf 'locked build touched target, lock, or staging\n' >&2
	exit 1
fi
rmdir "$target_lock"
expect_failure '' env FAKE_PARTIAL_FAIL=true FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target"
if [[ -e "$target" || -e "$target_lock" || -n "$(find "$clean_root" -maxdepth 1 -name '.published.*' -print -quit)" ]]; then
	printf 'failed build left output, lock, or staging behind\n' >&2
	exit 1
fi
env FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target" >/dev/null
if [[ -e "$target_lock" || -n "$(find "$clean_root" -maxdepth 1 -name '.published.*' -print -quit)" ]]; then
	printf 'successful build left lock or staging behind\n' >&2
	exit 1
fi
printf 'original asset\n' >"$target/asset.css"
cp -a "$target" "$suite_root/previous-output"
expect_failure '' env FAKE_PARTIAL_FAIL=true FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target"
if ! diff -r "$suite_root/previous-output" "$target" || [[ -e "$target_lock" ]] || \
	[[ -n "$(find "$clean_root" -maxdepth 1 -name '.published.*' -print -quit)" ]]; then
	printf 'failed build changed previous output or left staging behind\n' >&2
	exit 1
fi
env FAKE_INDEX=replacement FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target" >/dev/null
if [[ "$(cat "$target/index.html")" != replacement || -e "$target/asset.css" || -e "$target_lock" ]]; then
	printf 'successful build did not replace previous output\n' >&2
	exit 1
fi
cp -a "$target" "$suite_root/before-replacement"
expect_failure "MkDocs output changed during the build: $target" \
	env FAKE_REPLACE_TARGET="$target" FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target"
if [[ "$(cat "$target/marker")" != 'replacement marker' ]] || \
	[[ "$(find "$target" -mindepth 1 -maxdepth 1 | wc -l)" -ne 1 ]] || \
	! diff -r "$suite_root/before-replacement" "$target.original" || \
	[[ -n "$(find "$clean_root" -maxdepth 1 -name '.published.*' -print -quit)" ]]; then
	printf 'replacement changed or original, lock, or staging was not preserved correctly\n' >&2
	exit 1
fi
rm -rf -- "$target"
mv -- "$target.original" "$target"
expect_failure "MkDocs output changed during the build: $target" \
	env FAKE_REMOVE_TARGET="$target" FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target"
if [[ -e "$target" || -L "$target" || -n "$(find "$clean_root" -maxdepth 1 -name '.published.*' -print -quit)" ]]; then
	printf 'removed target reappeared or lock or staging remained\n' >&2
	exit 1
fi
cp -a "$suite_root/before-replacement" "$target"
appeared_target="$clean_root/appeared"
expect_failure "MkDocs output appeared during the build: $appeared_target" \
	env FAKE_APPEAR_TARGET="$appeared_target" FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$appeared_target"
if [[ ! -d "$appeared_target" || -n "$(find "$appeared_target" -mindepth 1 -print -quit)" || \
	-n "$(find "$clean_root" -maxdepth 1 -name '.appeared.*' -print -quit)" ]]; then
	printf 'appeared target holds nested output or staging remained\n' >&2
	exit 1
fi
if ((EUID == 0)); then
	printf 'skipping non-writable backup removal test as root\n'
else
	mkdir -p "$target/readonly"
	printf 'pinned\n' >"$target/readonly/file"
	chmod 0555 "$target/readonly"
	if removal_output="$(env FAKE_INDEX=published-after-readonly FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
		bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" --site-dir "$target" 2>&1)"; then
		chmod -R u+w "$clean_root"
		printf 'expected failure when earlier output cannot be removed\n' >&2
		exit 1
	fi
	chmod -R u+w "$clean_root"
	if [[ "$removal_output" != *"MkDocs could not remove earlier output:"* || \
		"$removal_output" == *"MkDocs could not restore output:"* ]]; then
		printf 'unexpected diagnostic on backup removal failure: %s\n' "$removal_output" >&2
		exit 1
	fi
	if [[ "$(cat "$target/index.html")" != published-after-readonly || \
		-n "$(find "$target" -mindepth 1 -maxdepth 1 -name '*.backup' -print -quit)" || \
		-e "$target/readonly" ]]; then
		printf 'failed backup removal corrupted published target\n' >&2
		exit 1
	fi
	rm -rf "$clean_root"/.published.*.backup
fi

FAKE_SECURITY_OPTIONS='["name=rootless"]' FAKE_DOCKER_LOG="$fake_log" DOCKER_BIN="$fake_docker" \
	bash "$repo_root/scripts/ci/mkdocs-build.sh" --root "$clean_root" >/dev/null
if grep -Fxq -- -u "$fake_log"; then
	printf 'rootless Docker must omit -u\n' >&2
	exit 1
fi

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
