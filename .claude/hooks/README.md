# Claude Code hooks — goenvoy

Three hooks enforce this repo's invariants at tool-call time. They run locally in your Claude Code session and refuse operations that violate the project rules.

## Hooks

### `guard-bash.sh` — PreToolUse / Bash
Blocks:
- `--no-verify` (disables pre-commit/push hooks)
- `--no-gpg-sign` (bypasses commit signing)
- `git push --force[-with-lease] ... main|master`
- `git reset --hard main|master`
- `rm -rf .git*`
- `rm -rf .workingdir*` (destroys plan + state)

### `guard-go-edit.sh` — PreToolUse / Edit|Write
Blocks Go-file edits that violate the pure-stdlib invariant or contain obvious client-lib footguns:
1. **Non-stdlib imports** — `github.com/...`, `gopkg.in/...`, `gitlab.com/...`, `bitbucket.org/...`. `golang.org/x/...` is allowed. (Enforces ADR-0001.)
2. **`InsecureSkipVerify: true`** — requires same-line `//nolint:gosec // <reason>`.
3. **`//nolint`** — requires same-line `// <reason>` justification.
4. **Live-API URLs in `*_test.go`** — tests must use `httptest.NewServer`; a denylist of known upstream hosts is blocked.

### `format-go-write.sh` — PostToolUse / Edit|Write
After every Go-file write, runs:
- `gofumpt -w <file>` (if installed)
- `gci write --skip-generated -s standard -s default -s 'prefix(github.com/golusoris/goenvoy)' <file>` (if installed)

Silently skips when either tool is missing — CI catches formatting regressions.

## Exemptions

- `guard-go-edit.sh` does **not** exempt `*_test.go` from the stdlib check — tests must also be stdlib-only.
- `golang.org/x/...` imports are allowed everywhere (they're effectively stdlib extensions).
- The live-API host denylist is small on purpose (four hosts today). Add to it when you see a regression.

## Smoke-testing a hook

Each hook reads a JSON payload on stdin. You can smoke-test without Claude Code:

```bash
echo '{"tool_input":{"command":"git push --force origin main"}}' | .claude/hooks/guard-bash.sh
# → exit 2, stderr: "blocked by ... force-push to main/master blocked..."

echo '{"tool_name":"Write","tool_input":{"file_path":"/tmp/x.go","content":"import \"github.com/foo/bar\""}}' | .claude/hooks/guard-go-edit.sh
# → exit 2, stderr: "non-stdlib import blocked..."

echo '{"tool_input":{"file_path":"/tmp/x.go"}}' | .claude/hooks/format-go-write.sh
# → exit 0 (no-op for missing file)
```

Exit code 0 = allowed. Exit code 2 = blocked (reason on stderr).

## Adding a new rule

1. Edit the relevant hook script.
2. Smoke-test with the snippet above.
3. Update this README with the new rule.
4. If the rule enforces an invariant also checkable in CI, add the equivalent linter/gate in [.golangci.yml](../../.golangci.yml) or [.github/workflows/ci.yml](../../.github/workflows/ci.yml) — hooks are best-effort; CI is the ground truth.
