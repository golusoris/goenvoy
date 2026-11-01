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

# rule 1: no non-stdlib imports (ADR-0001). Allowed: stdlib, golang.org/x/..., goenvoy's own modules.
# Why: go/parser reads the edited file as the compiler does; regex matching missed import forms.
# Why: tools/go is the dev-tool module (its go.mod; Makefile MODULES excludes it); ADR-0001 binds library modules.
case "$file_path" in
  tools/go/*|*/tools/go/*) ;;
  *)
    if command -v go >/dev/null 2>&1; then
      guard_src="$(dirname "$0")/importguard/main.go"
      guard_bin="${XDG_CACHE_HOME:-$HOME/.cache}/goenvoy-importguard/$(sha256sum "$guard_src" | cut -c1-16)"
      # Why: go run reports every non-zero exit as 1, so the binary is built once per source hash.
      [ -x "$guard_bin" ] || go build -o "$guard_bin" "$guard_src" >&2 || deny "importguard build failed"
      rc=0
      printf '%s' "$payload" | "$guard_bin" >&2 || rc=$?
      [ "$rc" = 2 ] && deny "non-stdlib import blocked (details above). goenvoy is pure stdlib (ADR-0001). If you truly need a dep, write an ADR first."
      [ "$rc" = 0 ] || deny "importguard exited $rc"
    fi
  ;;
esac

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
