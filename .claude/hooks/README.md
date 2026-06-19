# Claude Code hooks - goenvoy

Working hook scripts wired from [`../settings.json`](../settings.json). Each
script reads the Claude Code hook JSON envelope on stdin, inspects the tool
input, and either allows (exit 0), blocks with a rejection reason (exit 2), or
runs a post-action formatter.

The bar for adding rules here is: **would it save an edit cycle compared to
catching the same thing in CI?** `golangci-lint`, `gosec`, `govulncheck`, and
the module test matrix are authoritative. Hooks exist so the agent does not
spend a round trip on rules it already knows.

## Active Hooks

| Hook | Event | Matcher | Exit 2 triggers |
|---|---|---|---|
| [`guard-bash.sh`](guard-bash.sh) | PreToolUse | `Bash` | `--no-verify`, `--no-gpg-sign`, force-push to main/master, `reset --hard main/master`, `rm -rf .git`, `rm -rf .workingdir` |
| [`guard-go-edit.sh`](guard-go-edit.sh) | PreToolUse | `Edit\|Write` on `*.go` | Non-stdlib imports, unjustified `InsecureSkipVerify: true`, unjustified `//nolint`, live upstream URLs in tests |
| [`format-go-write.sh`](format-go-write.sh) | PostToolUse | `Edit\|Write` on `*.go` | Never blocks; runs `gofumpt -w` and `gci write` when those tools are on PATH |

## Path Notes

- `guard-go-edit.sh` does not exempt `*_test.go` from the stdlib check; tests
  must also stay pure stdlib.
- `golang.org/x/...` imports are allowed where needed for Go project tooling.
- The live-API host denylist is intentionally small. Add hosts when a real
  regression appears.

## Why These Rules

- **Pure stdlib** - ADR-0001 is load-bearing for the small client modules.
- **No silent TLS weakening** - `InsecureSkipVerify: true` needs a same-line
  `//nolint:gosec // <reason>` justification.
- **`//nolint` needs justification** - CI enforces this; catching it locally
  saves one push and wait cycle.
- **No live API tests** - goenvoy tests use `httptest`; secrets and network
  flakiness do not belong in module tests.

## Smoke-Testing A Hook

Each script reads the hook JSON envelope on stdin. To dry-run locally:

```bash
printf '%s' '{"tool_input":{"command":"git push --force origin main"}}' \
  | .claude/hooks/guard-bash.sh; echo "exit=$?"

printf '%s' '{"tool_name":"Write","tool_input":{"file_path":"/tmp/x.go","content":"import \"github.com/foo/bar\""}}' \
  | .claude/hooks/guard-go-edit.sh; echo "exit=$?"

printf '%s' '{"tool_input":{"file_path":"/tmp/x.go"}}' \
  | .claude/hooks/format-go-write.sh; echo "exit=$?"
```

A blocking hook returns exit code 2 and prints its reason to stderr; a passing
hook returns 0 silently.

## Settings Files - Tracked Vs Personal

- [`../settings.json`](../settings.json) - team-shared, checked in. This is
  where the hook wiring lives.
- `../settings.local.json` - personal, gitignored. User-specific overrides and
  API key paths go here.
- `../scheduled_tasks.lock` - runtime lockfile, gitignored.

If personal Claude files start showing up in `git status`, check
[`.gitignore`](../../.gitignore) before committing them.
