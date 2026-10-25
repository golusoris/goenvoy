<!-- markdownlint-disable MD013 -->
# golusoris/goenvoy Agent Operating Harness

<!-- praetor:head -->

Before concluding any turn:

```bash
make verify-all
```

`make verify-all` = repository-owned gate; adoption kept it unread + unexecuted. Steps live in `Makefile`: read there, never restate. Footer commands = project markers only.
Pass = exit 0. Signed Ed25519 Exit-0 receipt only from `praetorctl gate run`; report no receipt it did not mint. Fail -> SARIF diagnostic distillation (<= 1500 tokens).

## Core Directives & Invariants

Adopted check = coverage state + mechanism; else audit scanner; else `unknown`. Gates = (praetor `lefthook.yml`); edits not reflected.

| Invariant | Rule | Adopted check | On fail |
| :--- | :--- | :--- | :--- |
| **HISS-01** control flow | recursion prohibited; call graph = DAG; Go, C/C++: zero `goto` | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: Go `goto`, recursion + plain-function call cycles; Rust, Python direct recursion; C `goto` | new finding blocks commit, blocks push |
| **HISS-02** loops, I/O | scalar upper bound on every loop; explicit deadline on every I/O call; Go: I/O takes `context.Context` deadline | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: unbounded loop shapes in Go, C, Rust, Python; I/O deadlines unchecked | new finding blocks commit, blocks push |
| **HISS-03** memory | Go, Rust, C/C++: zero heap allocation in hot simulation/tick loops | unknown | advisory |
| **HISS-04** complexity | McCabe cyclomatic <= 10, cognitive <= 12, statements <= 45; func LOC <= 60 (audit ceiling) | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: function length only; other caps need repository linter | new finding blocks commit, blocks push |
| **HISS-05** scoping | declare every identifier in smallest lexical scope serving it | unknown | advisory |
| **HISS-06** concurrency | explicit scalar upper bound on every worker pool + concurrent fan-out | unknown | advisory |
| **HISS-07** errors | every error handled or wrapped with context; Go: zero unchecked `error` return; Rust: zero `.unwrap()` / `.expect()` outside tests | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: partial in Go, Rust, Python | new finding blocks commit, blocks push |
| **HISS-08** determinism | zero dynamic code execution (`eval` / `exec`); C/C++: zero banned libc (`gets` / `strcpy` / `sprintf`); a process signal's target is bounded above 1 (`killpg`, `kill(-pid)`, Go `syscall.Kill`) | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: C banned calls, Python `eval` / `exec`, Go `syscall.Kill` / `unix.Kill` | new finding blocks commit, blocks push |
| **HISS-09** reference safety | Go, Rust: `// SAFETY:` proof before every `unsafe` block | `praetorctl audit` HISS scan in lefthook pre-commit/pre-push: Go, Rust `unsafe` without proof; C, Python unchecked | new finding blocks commit, blocks push |
| **HISS-10** warnings | zero warnings: compiler, linter, format sweeps | `go vet` + `gofmt` in lefthook pre-commit: Go only | vet finding blocks commit |
| **HISS-11** supply chain | pinned lockfiles; zero floating tags; signed provenance | unknown | advisory |
| **HISS-12** secrets | zero credentials in Git history | unknown | advisory |
| **HISS-13** debt ratchet | recorded infractions never grow vs committed baseline | `praetorctl audit` ratchet vs `.standards-baseline.json` in lefthook pre-commit/pre-push | growth blocks commit, blocks push |
| **HISS-14** append-only ABI | public API append-only; breaking change = `!` subject + `Migration:` footer | unknown | advisory |
| **HISS-15** 3D testing | positive + negative + boundary tests, every public interface | unknown | advisory |
| **HISS-16** context integrity | single canonical `AGENTS.md`; vendor files compiled via `praetorctl compile-context` | `praetorctl compile-context --verify` in lefthook pre-commit | drift blocks commit |
| **HISS-17** state ledger | turn start `praetorctl state status`; turn end `praetorctl state sync .` | `praetorctl state sync .` in lefthook post-commit | none; records only |
| **HISS-18** CI efficiency | diff-aware gating via `praetorctl ci filter` | unknown | advisory |
| **HISS-19** reuse before writing | one behavior = one implementation; extend or call existing code | unknown | advisory |
| **HISS-20** replayable evidence | every enforcement claim backed by fixtures replayed both directions | unknown | advisory |
| **HISS-21** platform neutrality | gates, hooks, emitted templates run on Linux, macOS, Windows, or skip with stated reason | unknown | advisory |

## Operational Rules

1. **Act on verified state.** Read source files, run real commands before hypothesis or edit. Never guess flag names, library signatures, repo configuration from memory.

2. **Lead with output.** Direct answers, diffs, commands. No filler preamble, no "Based on", no restatement, no chatter.

3. **Context transpiler first.** Never edit `CLAUDE.md`, `.cursor/rules/hiss-invariants.mdc`, `.github/copilot-instructions.md`, `.windsurfrules`, `.gemini/GEMINI.md`, `.codex/rules.md` manually. All agent instruction updates -> `AGENTS.md`, then:

   ```bash
   praetorctl compile-context
   ```

   - `AGENTS.md` = agent-only text -> caveman (internal register). `praetorctl compile-context --verify` + `praetorctl audit` run caveman lint; findings fail gate; no opt-out. Check first: `praetorctl caveman check --kind=context AGENTS.md`.

4. **SARIF diagnostic distillation.** Compiler/linter errors -> distill to $\le 1,500$ tokens ($< 60$ lines): top 3 root-cause failures with file/line pointers; full SARIF logs -> ephemeral storage.

5. **No evasion.** Never attempt `--no-verify`, `LEFTHOOK=0`, or modifying `.git/hooks`. Hooks = local gate adoption installs. Adoption adds no server-side `praetorctl` gate run. Scaffolded CI: `.github/workflows/praetor-docs.yml` runs step `Stop on a draft pull request`, `node tools/markdownlint/verify.mjs`, step `Verify figures`; `.github/workflows/praetor-api.yml` runs step `Stop on a draft pull request`, `go run tools/apicompat/gate/main.go -base="$BASE"`.

6. **Anti-loop interception.** Same AST diff + error category repeats $\ge 3$ times -> halt immediately. Re-evaluate design; no micro-textual retries.

<!-- praetor:config -->

## Text Register

<!-- praetor:register:start -->
Register follows the audience, then the task label of your brief (`register:` in `.standards.yaml`; labels are the router's `target_tasks`).

| Register | Where | Form |
| :--- | :--- | :--- |
| social | forge: issues, PR bodies, review comments, commit bodies | `social-text` skill: BLUF, full sentences, scannable, enough and no more; conventional commit subject unchanged |
| docs | docs/, README, ADR bodies | complete without bloat: newcomer path first, expert reference after; every claim points at a file, command or test; no restated code |
| internal | briefs, agent-to-agent traffic, research fan-outs, workflow returns | `caveman` skill: fragments, no filler, verbatim code/paths/errors; facts, paths, commands, verdict |

- Task rows: social = commit_message_synthesis, waiver_signoff; docs = architecture_synthesis, function_docstrings; every other label and any unlabeled text = internal. Subagent launch brief: `caveman` brief shape with `task:` = routing label.
- Evidence above 58 lines or 1500 tokens leaves the message as a file under `.workingdir/evidence/`; return `evidence: <path> sha256:<12 hex> lines:<n>` and fetch it only when a decision needs it.
- An internal return carries verdict, changed paths, commands run, evidence pointers and open questions, nothing else.
<!-- praetor:register:end -->

<!-- praetor:tail -->

## Primary Verification Commands

```bash
# Fast local test suite
# Declared commands only; run them before claiming application verification.
'make' 'verify-all'

# Recompile and verify cross-agent context outputs
praetorctl compile-context --verify

# Audit repository against declared HISS standards
praetorctl audit

# Repository gate; steps live in Makefile
make verify-all
```

<!-- markdownlint-enable MD013 -->
<!-- markdownlint-disable MD025 -->
<!-- praetor:harness:end -->

---

# Agent guide — goenvoy

> Cross-tool context for Claude Code, Cursor, Aider, Codex, Continue, and other coding assistants.
> **Read before suggesting changes.** Read per-category `AGENTS.md` for area touched.

## What this repo is

`goenvoy` is **multi-module monorepo** of pure-stdlib Go HTTP-API clients — 69 modules covering 63+ services across arr-stack, metadata, download clients, media servers, anime. Each service is Go module at `github.com/golusoris/goenvoy/<category>/<service>`, independently versioned via `<category>/<service>/vX.Y.Z` tags.

## Hard rules

1. **Never add non-stdlib import.** Pure stdlib is load-bearing (ADR-0001). `net/http`, `encoding/json`, `encoding/xml`, `crypto/*`, `context`, `net/url`, `net/http/httptest` only. `depguard` in `.golangci.yml` enforces. If dependency considered, write ADR first.
2. **Every module exports `New(baseURL, apiKey string, opts ...Option) (*Client, error)`.** Exceptions go in module `AGENTS.md` (OAuth2 device-code, public-key-only APIs).
3. **Every I/O method takes `context.Context` as first argument.** No exceptions.
4. **Every module defines `APIError` struct implementing `error`.** Status code, message, response body.
5. **Every HTTP call closes response body.** `bodyclose` enforces.
6. **All errors wrap** — `fmt.Errorf("<module>: <op>: %w", err)`. `wrapcheck` enforces at boundaries.
7. **Never log secrets.** API keys / bearer tokens / refresh tokens must not appear in error messages, wrapped errors, or test fixtures.
8. **Every merged commit: 0 lint · 0 gosec · 0 govulncheck · race-green** — **per module**. `//nolint` requires same-line justification comment.
9. **Breaking public API requires `Migration:` footer** in commit body with before/after Go snippets. CI `apidiff` runs against last tag.

See [docs/principles.md](docs/principles.md) (adapted §2 contract) for full coding / security / testing rules.

## Repository layout

```
goenvoy/
├── arr/              # shared *arr types + base client
│   ├── sonarr/       # Sonarr v3 client
│   ├── radarr/       # Radarr v3 client
│   ├── lidarr/       # Lidarr client
│   ├── readarr/      # Readarr client
│   ├── whisparr/     # Whisparr v2 + v3 (Eros) client
│   ├── prowlarr/     # Prowlarr client
│   ├── bazarr/       # Bazarr client
│   ├── seerr/        # Seerr client
│   ├── autobrr/      # Autobrr client
│   ├── mylar/        # Mylar3 client
│   ├── jackett/      # Jackett client
│   └── nzbhydra/     # NZBHydra2 client
├── tools/            # adjacent automation utilities
│   └── flaresolverr/ # FlareSolverr Cloudflare-bypass client
├── metadata/         # shared metadata types (Rating, Image, Person, ...)
│   ├── video/        # TMDb, TheTVDB, Fanart.tv, OMDb, TVmaze, OpenSubtitles
│   ├── anime/        # AniList, Kitsu, AniDB, MAL
│   ├── music/        # MusicBrainz, Last.fm, Discogs, TheAudioDB, Spotify, Deezer
│   ├── tracking/     # Trakt, Simkl, Letterboxd, ListenBrainz
│   ├── adult/        # StashBox, TPDB
│   ├── book/         # Google Books, Open Library
│   └── game/         # IGDB, RAWG, Steam, MobyGames, SteamGridDB, RetroAchievements, ScreenScraper, Hasheous, LaunchBox
├── downloadclient/   # shared download types + Downloader interface
│   ├── qbit/         # qBittorrent WebUI client
│   ├── transmission/ # Transmission RPC client
│   ├── deluge/       # Deluge JSON-RPC client
│   ├── rtorrent/     # rTorrent XMLRPC client
│   ├── sabnzbd/      # SABnzbd client
│   └── nzbget/       # NZBGet JSON-RPC client
├── mediaserver/      # shared media server types
│   ├── plex/         # Plex Media Server client
│   ├── jellyfin/     # Jellyfin client
│   ├── emby/         # Emby client
│   ├── tautulli/     # Tautulli client
│   ├── audiobookshelf/ # Audiobookshelf client
│   ├── komga/        # Komga client
│   ├── navidrome/    # Navidrome client
│   ├── kavita/       # Kavita client
│   ├── stash/        # StashApp GraphQL client
│   └── tdarr/        # Tdarr client
└── anime/            # shared anime types
    └── shoko/        # Shoko Server client
```

## Common tasks

| Task | Claude Code skill / command |
| --- | --- |
| Add new service client | `/add-service-client <category> <service>` |
| Add method to existing client | `/add-service-method <module> <MethodName>` |
| Bump module version | `/bump-module <module> <level>` |
| Release module | `/release-module <module> <version>` |
| Audit + refresh pinned API docs | `/audit-service-docs` |

## CI gates (per module, in matrix)

Every PR runs, per module:

- `golangci-lint run` — 30+ linters, config at `.golangci.yml`.
- `gosec ./...` — SARIF uploaded to code-scanning.
- `govulncheck ./...`.
- `go test -race -count=1 -coverprofile=...` — coverage ≥ 70%.
- `apidiff` vs previous `<module>/vX.Y.Z` tag — breaking changes require `BREAKING CHANGE:` / `Migration:` footer.
- `go vet` + `go build`.

And repo-wide:

- PR title matches Conventional Commits 1.0.
- CodeQL (Go, `security-extended,security-and-quality`).
- OSSF Scorecard.

## Pinned upstream-API docs

`docs/upstream/<service>.md` records canonical API-doc URL, version module targets, and last-verified date. Check before suggesting new method — public docs may be ahead or behind.

## When in doubt

Read [docs/principles.md](docs/principles.md) for full contract, then per-category `AGENTS.md` for area touched.

<!-- Folded by praetorctl adopt from CLAUDE.md: text that file held and AGENTS.md lacked. -->

# Claude Code guide — goenvoy

> Claude Code-specific guide. For cross-tool conventions read [AGENTS.md](AGENTS.md) first; file extends it.

## Skills available

Located in `.claude/skills/`:

| Skill | When to use |
| --- | --- |
| `add-service-client` | Scaffold new API-client module from one-line prompt. |
| `add-service-method` | Add new typed method + test case to existing client. |
| `bump-module` | Bump module semver (major/minor/patch) and open PR. |
| `release-module` | Tag `<module>/vX.Y.Z`, trigger release.yml. |
| `audit-service-docs` | Refresh `docs/upstream/<service>.md` with current date + current URL. |

Invoke via `/<skill-name>` in Claude Code.

## Hooks active

Located in `.claude/hooks/`:

- **PreToolUse / Bash** — `guard-bash.sh` blocks `--no-verify`, `--no-gpg-sign`, force-push to main/master, `rm -rf .git`, `rm -rf .workingdir*`.
- **PreToolUse / Edit|Write** — `guard-go-edit.sh` blocks: non-stdlib imports (pure-stdlib invariant), `InsecureSkipVerify: true` without justified `//nolint:gosec`, unjustified `//nolint` directives, live-API URLs in `*_test.go`.
- **PostToolUse / Edit|Write** — `format-go-write.sh` runs `gofumpt -w` + `gci write -s standard -s default -s 'prefix(github.com/golusoris/goenvoy)'`.

## Tone

- Be terse. No preamble.
- When changing public API: write `Migration:` footer in commit body with before/after Go snippets.
- When adding method: always add table-driven test case + runnable godoc example.
- Never suggest adding dependency — goenvoy is pure stdlib by ADR-0001.

## Project principles

Read [docs/principles.md](docs/principles.md) first. Quick hitlist:

- Pure stdlib. No imports outside `net/http`, `encoding/json`, `encoding/xml`, `crypto/*`, `context`, `net/url`, `net/http/httptest`.
- `New(baseURL, apiKey string, opts ...Option) (*Client, error)` constructor shape, unless module `AGENTS.md` documents public-key-only or OAuth exception.
- Functional options (`Option` + `With*`).
- Every method takes `context.Context` first.
- Every module has `APIError`.
- Every response body is `defer resp.Body.Close()`-ed.
- `//nolint` needs same-line justification.

## Working agreements

- Use structured question popup when decision unclear and operationally risky; otherwise keep moving with repo established conventions.
- Update private `.workingdir/STATE.md` (local, never committed) as significant audit findings fixed, confirmed, or ruled out.
- Hardening pass should ship full set: code fixes, agent/tooling drift fixes, CI/local-gate parity, upstream-doc refreshes, verification summary.

## Don't

- Don't add external dependencies. Ever.
- Don't skip response-body close.
- Don't concatenate user input into URL paths without `url.PathEscape`.
- Don't silence linter without justification comment.
- Don't write multi-paragraph comments. One-line `// Why:` comments only.
- Don't create new markdown docs unless explicitly asked.
- Don't run tests against live APIs — `httptest` only.

## Per-commit doc-sync

When touching module:

- Update `<module>/AGENTS.md` if conventions changed.
- No hand edit of any `CHANGELOG.md`; release-please writes notes from the Conventional Commit subject, so the subject must describe the change.
- Update `docs/upstream/<service>.md` if upstream API surface moved.
