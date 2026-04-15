#!/usr/bin/env bash
# PreToolUse hook for Bash. Blocks trivially-wrong commands for this repo.
# Exit 2 = block with rejection reason on stderr.
set -euo pipefail
payload=$(cat)
if command -v jq >/dev/null 2>&1; then
  cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
else
  cmd=$(printf '%s' "$payload" | grep -oE '"command":"[^"]*"' | head -n1 | sed 's/^"command":"//; s/"$//')
fi
[ -z "$cmd" ] && exit 0

deny() { printf 'blocked by .claude/hooks/guard-bash.sh: %s\n' "$1" >&2; exit 2; }

case "$cmd" in
  *--no-verify*)   deny "'--no-verify' disables pre-commit/push hooks. Fix the failure instead." ;;
  *--no-gpg-sign*) deny "'--no-gpg-sign' bypasses commit signing. Not allowed." ;;
esac

if printf '%s' "$cmd" | grep -Eq 'git[[:space:]]+push[[:space:]].*--force(-with-lease)?\b.*\b(origin[[:space:]]+)?(main|master)\b'; then
  deny "force-push to main/master blocked. Create a PR or revert-commit instead."
fi

if printf '%s' "$cmd" | grep -Eq 'git[[:space:]]+reset[[:space:]].*--hard[[:space:]]+(origin/)?(main|master)\b'; then
  deny "'git reset --hard main' drops commits. Confirm with the user first."
fi

if printf '%s' "$cmd" | grep -Eq '\brm[[:space:]]+-[A-Za-z]*r[A-Za-z]*f?[A-Za-z]*[[:space:]]+(\./)?\.git(/[^[:space:]]*)?(\s|$)'; then
  deny "'rm -rf .git*' nukes the repo."
fi
if printf '%s' "$cmd" | grep -Eq '\brm[[:space:]]+-[A-Za-z]*r[A-Za-z]*f?[A-Za-z]*[[:space:]]+(\./)?\.workingdir2?(/|\s|$)'; then
  deny "'rm -rf .workingdir*' destroys plan + state."
fi
exit 0
