#!/usr/bin/env bash
# Rootless Docker maps container root to the caller; rootful Docker uses the caller UID and GID.

set -euo pipefail

repo_root="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
scan_root="$repo_root"
site_arg=""
while (($# > 0)); do
	case "$1" in
	--root)
		[[ $# -ge 2 ]] || {
			printf '%s requires a path\n' "$1" >&2
			exit 2
		}
		scan_root="$(CDPATH='' cd -- "$2" && pwd -P)"
		shift 2
		;;
	--site-dir)
		[[ $# -ge 2 ]] || {
			printf '%s requires a path\n' "$1" >&2
			exit 2
		}
		site_arg="$2"
		shift 2
		;;
	*)
		printf 'usage: %s [--root PATH] [--site-dir PATH]\n' "$0" >&2
		exit 2
		;;
	esac
done
readonly repo_root scan_root

# shellcheck source=/dev/null
. "$repo_root/tools/tool-versions.env"

if [[ ! -f "$scan_root/mkdocs.yml" || ! -d "$scan_root/docs" ]]; then
	printf 'MkDocs requires mkdocs.yml and docs/: %s\n' "$scan_root" >&2
	exit 1
fi
docs_dir="$(awk '/^["\047]?docs_dir["\047]?[[:space:]]*:/ { sub(/^[^:]*:[[:space:]]*/, ""); sub(/[[:space:]]+#.*/, ""); sub(/[[:space:]]+$/, ""); print "value:" $0 }' "$scan_root/mkdocs.yml")"
if [[ -n "$docs_dir" && "$docs_dir" != value:docs && "$docs_dir" != "value:'docs'" && "$docs_dir" != 'value:"docs"' ]]; then
	printf 'MkDocs refuses docs_dir: %s\n' "${docs_dir#value:}" >&2
	exit 1
fi
if symlink="$(find "$scan_root/docs" "$scan_root/mkdocs.yml" -type l -print -quit)" && \
	[[ -n "$symlink" ]]; then
	printf 'MkDocs build refuses symlink input: %s\n' "$symlink" >&2
	exit 1
fi
doc_count="$(find "$scan_root/docs" -type f | wc -l)"
if ((doc_count == 0 || doc_count > 8192)); then
	printf 'MkDocs document count outside bound 1..8192: %d\n' "$doc_count" >&2
	exit 1
fi

site_dir=""
target_site=""
backup_site=""
lock_dir=""
lock_held=false
had_target_site=false

# cleanup_site removes temporary output and restores a target if publication failed.
cleanup_site() {
	local status=$?
	if [[ -n "$backup_site" && -d "$backup_site" ]]; then
		if [[ -e "$target_site" || -L "$target_site" ]] || ! mv -- "$backup_site" "$target_site"; then
			printf 'MkDocs could not restore output: %s\n' "$target_site" >&2
			status=1
		fi
	fi
	if [[ -n "$site_dir" && -d "$site_dir" ]] && ! rm -rf -- "$site_dir"; then
		printf 'MkDocs could not remove temporary output: %s\n' "$site_dir" >&2
		status=1
	fi
	if [[ "$lock_held" == true && -n "$lock_dir" && -d "$lock_dir" ]] && ! rmdir -- "$lock_dir"; then
		printf 'MkDocs could not remove lock: %s\n' "$lock_dir" >&2
		status=1
	fi
	# Why: a return value from an EXIT trap never reaches the caller; exit does.
	exit "$status"
}
trap cleanup_site EXIT

if [[ -z "$site_arg" ]]; then
	site_dir="$(mktemp -d "${TMPDIR:-/tmp}/goenvoy-mkdocs.XXXXXX")"
else
	if [[ "$site_arg" == /* ]]; then
		target_site="$site_arg"
	else
		target_site="$scan_root/$site_arg"
	fi
	IFS=/ read -r -a components <<<"$target_site"
	if ((${#components[@]} > 64)); then
		printf 'MkDocs output exceeds 64 path components\n' >&2
		exit 1
	fi
	resolved=""
	for component in "${components[@]}"; do
		case "$component" in
		""|.) continue ;;
		..) resolved="${resolved%/*}" ;;
		*) resolved="$resolved/$component" ;;
		esac
		if [[ -L "$resolved" ]]; then
			printf 'MkDocs output refuses symlink: %s\n' "$resolved" >&2
			exit 1
		fi
	done
	target_site="${resolved:-/}"
	if [[ "$target_site" == "$scan_root" || "$target_site" == "$scan_root/docs" || "$target_site" == "$scan_root/docs/"* || "$scan_root" == "$target_site/"* || "$target_site" == / ]]; then
		printf 'MkDocs output overlaps source: %s\n' "$target_site" >&2
		exit 1
	fi
	if [[ -d "$target_site" && -n "$(find "$target_site" -mindepth 1 -maxdepth 1 -print -quit)" ]] && \
		{ [[ ! -f "$target_site/index.html" ]] || [[ -e "$target_site/mkdocs.yml" || -e "$target_site/.git" ]]; }; then
		printf 'MkDocs output refuses nonempty directory: %s\n' "$target_site" >&2
		exit 1
	fi
	if [[ -e "$target_site" && ! -d "$target_site" ]]; then
		printf 'MkDocs output is not a directory: %s\n' "$target_site" >&2
		exit 1
	fi
	mkdir -p -- "$(dirname -- "$target_site")"
	lock_dir="$(dirname -- "$target_site")/.$(basename -- "$target_site").lock"
	if ! mkdir -- "$lock_dir" 2>/dev/null; then
		printf 'MkDocs output is locked by another build: %s\n' "$lock_dir" >&2
		exit 1
	fi
	lock_held=true
	if [[ -d "$target_site" ]]; then
		had_target_site=true
	fi
	site_dir="$(mktemp -d "$(dirname -- "$target_site")/.$(basename -- "$target_site").XXXXXX")"
fi
readonly site_dir target_site lock_dir lock_held had_target_site

docker_bin="${DOCKER_BIN:-}"
if [[ -z "$docker_bin" ]] && command -v docker >/dev/null; then
	docker_bin="$(command -v docker)"
fi
if [[ -z "$docker_bin" || ! -x "$docker_bin" ]]; then
	printf 'Docker is required for the digest-pinned MkDocs build\n' >&2
	exit 1
fi

image="squidfunk/mkdocs-material:${MKDOCS_MATERIAL_VERSION}@${MKDOCS_MATERIAL_IMAGE_DIGEST}"
security_options="$("$docker_bin" info --format '{{json .SecurityOptions}}')"
docker_args=(run --rm --network none --read-only --cap-drop ALL --security-opt no-new-privileges)
if [[ "$security_options" != *rootless* ]]; then
	docker_args+=(-u "$(id -u):$(id -g)")
fi
docker_args+=(--tmpfs '/tmp:rw,noexec,nosuid,size=64m' -e NO_MKDOCS_2_WARNING=1
	-v "$scan_root:/docs:ro" -v "$site_dir:/site" -w /docs
	"$image" build --strict --config-file /docs/mkdocs.yml --site-dir /site)
"$docker_bin" "${docker_args[@]}"

if [[ ! -f "$site_dir/index.html" ]]; then
	printf 'MkDocs build produced no index.html\n' >&2
	exit 1
fi
if [[ -n "$target_site" ]]; then
	chmod 0755 -- "$site_dir"
	if [[ "$had_target_site" == true && -d "$target_site" ]]; then
		backup_site="${site_dir}.backup"
		mv -- "$target_site" "$backup_site"
	fi
	if [[ -e "$target_site" || -L "$target_site" ]]; then
		printf 'MkDocs output appeared during the build: %s\n' "$target_site" >&2
		exit 1
	fi
	mv -- "$site_dir" "$target_site"
	if [[ -n "$backup_site" ]]; then
		old_site="$backup_site"
		backup_site=""
		if ! rm -rf -- "$old_site"; then
			printf 'MkDocs could not remove earlier output: %s\n' "$old_site" >&2
			exit 1
		fi
	fi
fi
printf 'MkDocs Material %s strict build passed for %d documents\n' \
	"$MKDOCS_MATERIAL_VERSION" "$doc_count"
