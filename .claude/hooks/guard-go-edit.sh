#!/usr/bin/env bash
# PreToolUse hook for Edit / Write on *.go files.
# Enforces the pure-stdlib invariant and common client-lib footguns.
set -euo pipefail
payload=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

file_path=$(printf '%s' "$payload" | jq -r '.tool_input.file_path // ""')
tool_name=$(printf '%s' "$payload" | jq -r '.tool_name // ""')
[ -z "$file_path" ] && exit 0
[[ "$file_path" != *.go ]] && exit 0

case "$tool_name" in
  Write) content=$(printf '%s' "$payload" | jq -r '.tool_input.content // ""') ;;
  Edit)  content=$(printf '%s' "$payload" | jq -r '.tool_input.new_string // ""') ;;
  *)     exit 0 ;;
esac
[ -z "$content" ] && exit 0

deny() { printf 'blocked by .claude/hooks/guard-go-edit.sh (%s):\n  %s\n' "$file_path" "$1" >&2; exit 2; }

# rule 1: no non-stdlib imports (ADR-0001). golang.org/x/... allowed.
if printf '%s' "$content" | grep -E '^[[:space:]]*"(github\.com|gopkg\.in|gitlab\.com|bitbucket\.org)/' >/dev/null; then
  deny "non-stdlib import blocked — goenvoy is pure stdlib (ADR-0001). If you truly need a dep, write an ADR first."
fi

# rule 2: InsecureSkipVerify without justification
if printf '%s' "$content" | grep -E 'InsecureSkipVerify:[[:space:]]*true' >/dev/null; then
  if ! printf '%s' "$content" | grep -E 'InsecureSkipVerify:[[:space:]]*true.*//[[:space:]]*nolint:gosec.*//' >/dev/null; then
    deny "InsecureSkipVerify: true requires a same-line '//nolint:gosec // <reason>' justification."
  fi
fi

# rule 3: //nolint without justification
if printf '%s' "$content" | grep -E '//[[:space:]]*nolint(:[[:alnum:],_-]+)?[[:space:]]*$' >/dev/null; then
  deny "//nolint needs a same-line justification, e.g. '//nolint:errcheck // defer Close, error surfaced'."
fi

# rule 4: live-API host in test files
case "$file_path" in
  *_test.go)
    if printf '%s' "$content" | grep -E '"https?://(api\.tmdb\.org|api\.trakt\.tv|api\.themoviedb\.org|anilist\.co|graphql\.anilist\.co|kitsu\.io|api\.thetvdb\.com)' >/dev/null; then
      deny "live-API URL in test. Use httptest.NewServer — goenvoy tests MUST NOT hit real APIs."
    fi
  ;;
esac

exit 0
