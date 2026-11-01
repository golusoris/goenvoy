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
# Why: Edit payloads are often bare import-block fragments, so every import-spec-shaped line counts.
# Why: a complete file (package clause) is read only up to its first declaration; imports precede
# every declaration, so raw-string fixtures further down are never taken for imports.
whole=0
# Why: tools/go is the dev-tool module (its go.mod; Makefile MODULES excludes it); ADR-0001 binds library modules.
case "$file_path" in tools/go/*|*/tools/go/*) whole=skip ;; esac
[ "$whole" = 0 ] && grep -qE '^package[[:space:]]+[A-Za-z_]' <<<"$content" && whole=1
in_block=0
foreign=""
while [ "$whole" != skip ] && IFS= read -r line; do
  line="${line#"${line%%[![:space:]]*}"}"
  if [ "$whole" = 1 ]; then
    [[ $line =~ ^(func|type|var|const)([[:space:]]|\(|$) ]] && break
    [[ $line =~ ^import[[:space:]]*\([[:space:]]*$ ]] && { in_block=1; continue; }
    [[ $in_block = 1 && $line == \)* ]] && { in_block=0; continue; }
    [[ $in_block = 0 && ! $line =~ ^import[[:space:]] ]] && continue
  fi
  [[ $line =~ ^import[[:space:]]+(.*)$ ]] && line="${BASH_REMATCH[1]}"
  [[ $line =~ ^(([A-Za-z_][A-Za-z0-9_]*|\.)[[:space:]]+)?\"([^\"]+)\"[[:space:]]*(//.*)?$ ]] || continue
  case "${BASH_REMATCH[2]}" in return|go|defer|case|else) continue ;; esac
  path="${BASH_REMATCH[3]}"
  case "$path" in github.com/golusoris/goenvoy/*|golang.org/x/*) continue ;; esac
  # Why: stdlib paths never carry a dot in their first element; module paths always do.
  [[ ${path%%/*} == *.* ]] && foreign+="$path "
done <<<"$content"
if [ -n "$foreign" ]; then
  deny "non-stdlib import blocked (${foreign% }) — goenvoy is pure stdlib (ADR-0001). If you truly need a dep, write an ADR first."
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
