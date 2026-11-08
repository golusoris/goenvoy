#!/usr/bin/env bash
# Why: planted test cases and repository check for release-please configuration.
set -euo pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
  echo "release-config_test.sh: needs bash >= 4 (mapfile); macOS: install a current bash" >&2
  exit 2
fi

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$HERE/.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# Why: assert command fails and error message contains needle.
expect_refusal() {
  local name="$1" needle="$2" out
  shift 2
  if out="$("$@" 2>&1)"; then
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

# Why: assert command succeeds.
expect_pass() {
  local name="$1" out
  shift 1
  if ! out="$("$@" 2>&1)"; then
    printf 'FAIL %s: command failed\n%s\n' "$name" "$out" >&2
    exit 1
  fi
  echo "ok   $name: $out"
}

# Why: stage tracked go.mod files and generator into an isolated temporary git repo.
setup_generator_dir() {
  local target="$1"
  local -a mod_files
  local f
  mkdir -p "$target/tools"
  cp "$REPO_ROOT/tools/gen-release-please.sh" "$target/tools/gen-release-please.sh"
  mapfile -t mod_files < <(git -C "$REPO_ROOT" ls-files '*go.mod')
  for f in "${mod_files[@]}"; do
    mkdir -p "$target/$(dirname "$f")"
    cp "$REPO_ROOT/$f" "$target/$f"
  done
  git -C "$target" init -q
  git -C "$target" add -A
  (
    cd "$target"
    bash tools/gen-release-please.sh >/dev/null 2>&1
  )
}

# Why: verify config matches generator output when gen_dir is provided.
check_generator_match() {
  local config="$1"
  local gen_dir="${2:-}"

  if [ -n "$gen_dir" ]; then
    if [ ! -d "$gen_dir" ] || [ ! -f "$gen_dir/tools/gen-release-please.sh" ]; then
      printf 'FAIL %s: generator directory invalid: %s\n' "$config" "$gen_dir" >&2
      return 1
    fi
    (
      cd "$gen_dir"
      bash tools/gen-release-please.sh >/dev/null 2>&1
    )
    if ! cmp -s "$gen_dir/release-please-config.json" "$config"; then
      printf 'FAIL %s: does not match generator output; edit tools/gen-release-please.sh to update\n' "$config" >&2
      return 1
    fi
  fi
  return 0
}

# Why: verify pull-request-header and separate-pull-requests fields are valid.
check_config_structure() {
  local config="$1"
  local header sep_pr

  if ! jq -e 'has("pull-request-header") and ."pull-request-header" != null' "$config" >/dev/null 2>&1; then
    printf 'FAIL %s: missing pull-request-header\n' "$config" >&2
    return 1
  fi

  header="$(jq -r '."pull-request-header"' "$config")"
  if [ -z "$header" ]; then
    printf 'FAIL %s: pull-request-header is empty\n' "$config" >&2
    return 1
  fi

  sep_pr="$(jq -r '."separate-pull-requests"' "$config")"
  if [ "$sep_pr" != "false" ]; then
    printf 'FAIL %s: separate-pull-requests must be false (got: %s)\n' "$config" "$sep_pr" >&2
    return 1
  fi
  return 0
}

# Why: verify pull-request-header contains all top-level headings from PR template.
check_header_headings() {
  local config="$1"
  local template="$REPO_ROOT/.github/pull_request_template.md"
  local header req line found
  local -a required_headings header_lines

  if [ ! -f "$template" ]; then
    printf 'FAIL %s: PR template not found at %s\n' "$config" "$template" >&2
    return 1
  fi

  mapfile -t required_headings < <(grep '^## ' "$template")
  if [ ${#required_headings[@]} -eq 0 ]; then
    printf 'FAIL %s: no headings found in PR template\n' "$config" >&2
    return 1
  fi

  header="$(jq -r '."pull-request-header"' "$config")"
  mapfile -t header_lines < <(printf '%s\n' "$header")

  for req in "${required_headings[@]}"; do
    req="${req%$'\r'}"
    found=0
    for line in "${header_lines[@]}"; do
      line="${line%$'\r'}"
      if [ "$line" = "$req" ]; then
        found=1
        break
      fi
    done
    if [ "$found" -ne 1 ]; then
      printf 'FAIL %s: pull-request-header missing required heading: %s\n' "$config" "$req" >&2
      return 1
    fi
  done
  return 0
}

# Why: verify pull-request-header does not contain forbidden separator lines.
check_header_separator() {
  local config="$1"
  local header line
  local -a header_lines

  header="$(jq -r '."pull-request-header"' "$config")"
  mapfile -t header_lines < <(printf '%s\n' "$header")

  for line in "${header_lines[@]}"; do
    line="${line%$'\r'}"
    if [ "$line" = "---" ]; then
      printf 'FAIL %s: pull-request-header contains forbidden "---" line\n' "$config" >&2
      return 1
    fi
  done
  return 0
}

# Why: verify pull-request-header contains no pre-ticked checklist boxes.
check_header_ticked() {
  local config="$1"
  local header

  header="$(jq -r '."pull-request-header"' "$config")"
  case "$header" in
    *"- [x]"*|*"- [X]"*)
      printf 'FAIL %s: pull-request-header contains a ticked box; leave checklist items for CI\n' "$config" >&2
      return 1
      ;;
  esac
  return 0
}

# Why: verify release-please config structure, PR template headings, and generator fidelity.
check_release_config() {
  local config="$1"
  local gen_dir="${2:-}"

  if [ ! -f "$config" ]; then
    printf 'FAIL %s: config file not found\n' "$config" >&2
    return 1
  fi

  check_config_structure "$config" || return 1
  check_header_headings "$config" || return 1
  check_header_separator "$config" || return 1
  check_header_ticked "$config" || return 1
  check_generator_match "$config" "$gen_dir" || return 1

  printf 'release-please config matches generator and PR template\n'
}

# Why: verify generator does not include untracked modules in config or manifest.
check_untracked_module() {
  local gen_dir="$1"
  local extra_mod="$2"
  (
    cd "$gen_dir"
    bash tools/gen-release-please.sh >/dev/null 2>&1
  ) || {
    printf 'FAIL %s: generator failed\n' "case_g" >&2
    return 1
  }
  if grep -q "\"$extra_mod\"" "$gen_dir/release-please-config.json" "$gen_dir/.release-please-manifest.json"; then
    printf 'FAIL %s: generator output includes untracked module: %s\n' "case_g" "$extra_mod" >&2
    return 1
  fi
  printf 'untracked module %s absent from generator output\n' "$extra_mod"
}

# Why: verify generator fails atomically without modifying existing files on tag error.
check_atomic_failure() {
  local gen_dir="$1"
  local gen_rc=0
  (
    cd "$gen_dir"
    PATH="$gen_dir:$PATH" bash tools/gen-release-please.sh >/dev/null 2>&1
  ) || gen_rc=$?
  if [ "$gen_rc" -eq 0 ]; then
    printf 'FAIL %s: generator unexpectedly succeeded despite failing git tag\n' "case_h" >&2
    return 1
  fi
  if ! cmp -s "$gen_dir/release-please-config.json" "$gen_dir/config.before"; then
    printf 'FAIL %s: release-please-config.json modified on failure\n' "case_h" >&2
    return 1
  fi
  if ! cmp -s "$gen_dir/.release-please-manifest.json" "$gen_dir/manifest.before"; then
    printf 'FAIL %s: .release-please-manifest.json modified on failure\n' "case_h" >&2
    return 1
  fi
  printf 'generator failed atomically (rc=%d), config and manifest unmodified\n' "$gen_rc"
}

GEN_DIR="$WORK/gen"
setup_generator_dir "$GEN_DIR"
BASE_CONFIG="$GEN_DIR/release-please-config.json"

# Why: case a: config without pull-request-header must be refused.
jq 'del(."pull-request-header")' "$BASE_CONFIG" > "$WORK/case_a.json"
expect_refusal "case_a: config without pull-request-header" "missing pull-request-header" check_release_config "$WORK/case_a.json"

# Why: case b: header missing ## Checklist heading must be refused.
jq '."pull-request-header" |= gsub("## Checklist"; "## Notes")' "$BASE_CONFIG" > "$WORK/case_b.json"
expect_refusal "case_b: header missing ## Checklist" "## Checklist" check_release_config "$WORK/case_b.json"

# Why: case c: header containing a '---' line must be refused.
jq '."pull-request-header" += "\n---\n"' "$BASE_CONFIG" > "$WORK/case_c.json"
expect_refusal "case_c: header with a '---' line" "---" check_release_config "$WORK/case_c.json"

# Why: case d: separate-pull-requests set to true must be refused.
jq '."separate-pull-requests" = true' "$BASE_CONFIG" > "$WORK/case_d.json"
expect_refusal "case_d: separate-pull-requests is true" "separate-pull-requests" check_release_config "$WORK/case_d.json"

# Why: case e: committed config differing from generator output by one byte must be refused.
sed 's/release-type/Release-type/' "$BASE_CONFIG" > "$WORK/case_e.json"
expect_refusal "case_e: committed config differing from generator output" "tools/gen-release-please.sh" check_release_config "$WORK/case_e.json" "$GEN_DIR"

# Why: case f: header containing a ticked box (- [x]) must be refused.
jq '."pull-request-header" += "\n- [x] ticked checklist item\n"' "$BASE_CONFIG" > "$WORK/case_f.json"
expect_refusal "case_f: header with a ticked box" "ticked" check_release_config "$WORK/case_f.json"

# Why: case g: untracked go.mod in working tree must not enter generator output.
CASE_G_DIR="$WORK/case_g"
setup_generator_dir "$CASE_G_DIR"
mkdir -p "$CASE_G_DIR/extra/mod"
touch "$CASE_G_DIR/extra/mod/go.mod"
expect_pass "case_g: generator ignores untracked module" check_untracked_module "$CASE_G_DIR" "extra/mod"

# Why: case h: failing git tag must fail atomically without touching config or manifest.
CASE_H_DIR="$WORK/case_h"
setup_generator_dir "$CASE_H_DIR"
cp "$REPO_ROOT/release-please-config.json" "$CASE_H_DIR/release-please-config.json"
cp "$REPO_ROOT/.release-please-manifest.json" "$CASE_H_DIR/.release-please-manifest.json"
cp "$CASE_H_DIR/release-please-config.json" "$CASE_H_DIR/config.before"
cp "$CASE_H_DIR/.release-please-manifest.json" "$CASE_H_DIR/manifest.before"

REAL_GIT="$(command -v git)"
cat <<EOFSTUB > "$CASE_H_DIR/git"
#!/usr/bin/env bash
if [ "\${1:-}" = "tag" ]; then
  exit 1
fi
exec "$REAL_GIT" "\$@"
EOFSTUB
chmod +x "$CASE_H_DIR/git"

expect_pass "case_h: generator fails atomically on tag error" check_atomic_failure "$CASE_H_DIR"

# Why: positive case: committed repository configuration matches generator and PR template.
expect_pass "repo: committed config matches generator and PR template" check_release_config "$REPO_ROOT/release-please-config.json" "$GEN_DIR"

echo "release-config: all planted cases pass"
