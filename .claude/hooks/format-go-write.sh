#!/usr/bin/env bash
# PostToolUse hook for Edit / Write on *.go files. Auto-formats with gofumpt + gci.
set -uo pipefail
payload=$(cat)
command -v jq >/dev/null 2>&1 || exit 0
file_path=$(printf '%s' "$payload" | jq -r '.tool_input.file_path // ""')
[ -z "$file_path" ] && exit 0
[[ "$file_path" != *.go ]] && exit 0
[ ! -f "$file_path" ] && exit 0

if command -v gofumpt >/dev/null 2>&1; then gofumpt -w "$file_path" 2>/dev/null || true; fi
if command -v gci >/dev/null 2>&1; then
  gci write --skip-generated -s standard -s default -s 'prefix(github.com/golusoris/goenvoy)' "$file_path" 2>/dev/null || true
fi
exit 0
