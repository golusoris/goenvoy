#!/usr/bin/env bash
# Regenerate release-please-config.json and .release-please-manifest.json
# from the current set of go.mod files + existing git tags.
#
# Run from repo root: tools/gen-release-please.sh
set -euo pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4 ]; then
  echo "gen-release-please.sh: needs bash >= 4 (mapfile); macOS: install a current bash" >&2
  exit 2
fi

CONFIG=release-please-config.json
MANIFEST=.release-please-manifest.json

# Why: sort in byte order so CI and local runs agree.
mapfile -t MODS < <(git ls-files -- '*go.mod' ':(exclude)tools/go/*' | while IFS= read -r f; do dirname "$f"; done | LC_ALL=C sort)

TMP_CONFIG=$(mktemp .release-please-config.XXXXXX)
TMP_MANIFEST=$(mktemp .release-please-manifest.XXXXXX)
trap 'rm -f "$TMP_CONFIG" "$TMP_MANIFEST"' EXIT

# --- config ---
{
  cat <<'EOF'
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "go",
  "bump-minor-pre-major": true,
  "bump-patch-for-minor-pre-major": false,
  "include-v-in-tag": true,
  "separate-pull-requests": false,
  "tag-separator": "/",
  "pull-request-header": "## Description\n\nRelease pull request opened by release-please. The release notes below list each module version and the Conventional Commit subjects it ships; release-please wrote the version bumps and changelogs from those subjects.\n\n## Module(s) Affected\n\nThe modules named in the release notes below.\n\n## Type of Change\n\nRelease: version bumps and changelogs only.\n\n## Checklist\n\n- [ ] Required status checks pass (`CI success`, `Go API Compatibility` and the rest of the `main` ruleset); the ruleset blocks the merge until they do.",
  "changelog-sections": [
    { "type": "feat", "section": "Features" },
    { "type": "fix", "section": "Bug Fixes" },
    { "type": "perf", "section": "Performance" },
    { "type": "refactor", "section": "Code Refactoring" },
    { "type": "deps", "section": "Dependencies" },
    { "type": "docs", "section": "Documentation", "hidden": true },
    { "type": "test", "section": "Tests", "hidden": true },
    { "type": "ci", "section": "CI", "hidden": true },
    { "type": "build", "section": "Build", "hidden": true },
    { "type": "chore", "section": "Chores", "hidden": true },
    { "type": "style", "section": "Style", "hidden": true },
    { "type": "revert", "section": "Reverts", "hidden": true }
  ],
  "packages": {
EOF

  last=$((${#MODS[@]} - 1))
  for i in "${!MODS[@]}"; do
    mod="${MODS[$i]}"
    sep=","
    [ "$i" -eq "$last" ] && sep=""
    printf '    "%s": { "package-name": "goenvoy/%s", "component": "%s" }%s\n' \
      "$mod" "$mod" "$mod" "$sep"
  done

  cat <<'EOF'
  }
}
EOF
} > "$TMP_CONFIG"

# --- manifest (latest tag per module, stripped of leading 'v') ---
{
  echo '{'
  last=$((${#MODS[@]} - 1))
  for i in "${!MODS[@]}"; do
    mod="${MODS[$i]}"
    sep=","
    [ "$i" -eq "$last" ] && sep=""
    # Why: avoid pipe to head under pipefail; take first line via parameter expansion.
    latest=$(git tag --list "${mod}/v[0-9]*" --sort=-v:refname)
    latest=${latest%%$'\n'*}
    if [ -n "$latest" ]; then
      ver="${latest##*/v}"
    else
      ver="0.0.0"
    fi
    printf '  "%s": "%s"%s\n' "$mod" "$ver" "$sep"
  done
  echo '}'
} > "$TMP_MANIFEST"

mv "$TMP_CONFIG" "$CONFIG"
mv "$TMP_MANIFEST" "$MANIFEST"

echo "wrote $CONFIG and $MANIFEST (${#MODS[@]} packages)"
