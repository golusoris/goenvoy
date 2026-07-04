# Session state — goenvoy

> Persistent state across workstations and AI sessions. Updated as significant changes happen.
> **Convention (per issue #12): write back to this file immediately as work lands — no batching.**
> Last update: 2026-07-04 (verified phases 2–6 landed; marked them Done with landing evidence).

## Standards rollout — `.goenvoy2.0/`

Phased adoption of golusoris standards — **complete**. Plan in [`.goenvoy2.0/`](../.goenvoy2.0/), checklist in [`.goenvoy2.0/10-rollout-checklist.md`](../.goenvoy2.0/10-rollout-checklist.md).

| Phase | Title | Status |
|---|---|---|
| 1 | Governance docs + editor config | Done — root `AGENTS.md`/`CLAUDE.md`/`.editorconfig`/`.markdownlintignore` (commit `0fbb62d`, #6) |
| 2 | Principles + ADRs + working dir | Done — `.workingdir/PRINCIPLES.md` + `docs/adr/0001`–`0009` (commit `0fbb62d`, #6) |
| 3 | Lint + tooling baseline | Done — `.golangci.yml` (30+ linters incl. depguard stdlib gate) + `Makefile` gates (lint v2 baseline, commit `0fbb62d`, #6) |
| 4 | CI hardening (CodeQL, Scorecard, apidiff) | Done — `.github/workflows/codeql.yml` + `scorecard.yml` (commit `5884791`); apidiff gate in `ci.yml` fixed 2026-06-19 |
| 5 | Release + supply-chain hardening | Done — release-please (`release-please.yml`, `release-please-config.json`, `.release-please-manifest.json`) (commit `ffe809a`); live via `chore: release main (#68)` `53be042` |
| 6 | Claude hooks + skills + per-module AGENTS | Done — `.claude/hooks/` (3 guards + formatter) + 5 skills + 70 `AGENTS.md` (root + 69 modules) (commit `4beed3d`) |

## 2026-07-04 Notes

- Refreshed this stale STATE.md (GOE-2). The phase table had marked phases 2–6
  "In progress" since 2026-06-19, but all six phases have LANDED in-repo. Verified
  each artifact exists and captured the introducing commit as evidence (table above):
  - Phase 2: `.workingdir/PRINCIPLES.md` + `docs/adr/0001`–`0009` (nine ADRs, one
    past the planned `0008`; `0009-local-replace-directives.md` also present).
  - Phase 3: `.golangci.yml` (30+ linters, depguard stdlib-only gate) + `Makefile` gates.
  - Phase 4: `codeql.yml` + `scorecard.yml` workflows; apidiff gate fixed 2026-06-19.
  - Phase 5: release-please trio present and live (last release `53be042`, #68).
  - Phase 6: `.claude/hooks/` (guard-bash, guard-go-edit, format-go-write) + 5 skills
    (add-service-client, add-service-method, bump-module, release-module,
    audit-service-docs) + 70 `AGENTS.md` (root + all 69 modules).
- Corrected the Phase 1 row: its prior evidence hash `0e4352b` resolves in no ref
  (dangling). The phase-1 governance files (root `AGENTS.md`/`CLAUDE.md`/
  `.editorconfig`/`.markdownlintignore`) actually landed in `0fbb62d` (#6).
- No re-implementation done — everything already present. No genuine gap surfaced.
- Adopted the immediate-writeback convention (issue #12): update this file as work
  lands, not in batches.

## 2026-06-14 Notes

- Merged `origin/main` into local `main`; kept the moved Letterboxd module at
  `metadata/tracking/letterboxd`.
- Verified installed Go is `go1.26.4`, matching the current stable Go patch line.
- Bumped every module directive and tracked templates from `1.26.1` to `1.26.4`.
- Compared automation against `/home/kilian/dev/golusoris` and began aligning
  Renovate policy, pinned local tools, standalone `gosec`, coverage parity, and
  release-check safety.
- Fixed nil-parameter search panics in TPDB, Google Books, and OpenSubtitles;
  OpenSubtitles parent search filters are now encoded.
- Refreshed stale upstream-doc URLs for Autobrr, Bazarr, Kavita, Komga,
  TheAudioDB, OpenSubtitles, LaunchBox, and TPDB; fixed metadata service
  `AGENTS.md` detail links to point back to root `docs/upstream`.

## 2026-06-19 Notes

- Verified the still-uncommitted 2026-06-14 hardening pass is fully green against
  the pinned CI gates across all 69 modules: build + vet + `go test -race` +
  `golangci-lint` v2.12.2 + `gosec` v2.27.1 + `govulncheck`; `coverage-check-all`
  passes (no module below threshold). The new `arr/whisparr/whisparr_extra_test.go`
  brings whisparr to 76.7%.
- Confirmed the gosec G704/G709 taint-analysis rules are not enabled by default in
  the pinned `gosec` v2.27.1, so the listenbrainz `//nolint:gosec` removal is clean
  and no `#nosec` replacement is required. A divergent locally-built gosec does flag
  them — always verify against the pinned version.
- Version currency: Go (go1.26.4), golangci-lint (v2.12.2), and gosec (v2.27.1) are
  already latest. Bumped `govulncheck` v1.3.0 → v1.4.0 (Makefile + ci.yml); verified
  clean.
- GitHub state: folded `actions/checkout` v6 → v7 into the workflows (incl. the
  orphaned `de0fac2e` SHA in codeql.yml) and closed Renovate PR #69 as superseded.
  The Renovate Dependency Dashboard (#62) listed no unopened version bumps beyond
  #69. release-please PR #68 remains pending.
- Pushing surfaced that local `main` was 11 commits ahead of origin with a
  half-finished major-version migration, and that `main` had been CI-red since
  2026-06-14. Two inherited problems, both now fixed:
  1. The apidiff CI gate never worked — it called `apidiff -m IMPORT DIR > file`
     (a no-op for this apidiff version), so every tagged module failed the
     empty-fingerprint guard. Fixed to `apidiff -m -w FILE MODULE` from inside each
     module dir + `apidiff -incompatible -m` output-based detection (apidiff exits
     0 even on incompatible changes).
  2. The `New() → (*Client, error)` constructor sweep broke 22 already-released
     modules on the same major path. Bumped each to the next major (18 × v1→/v2;
     arr/mylar, mediaserver/emby, mediaserver/jellyfin, mediaserver/tdarr × v2→/v3).
     All 22 build/vet/test green; the apidiff gate now reports 0 breaking.
- Verified locally before push: apidiff audit (0 breaking, 28 path-change skips),
  lint on migrated modules, gofumpt clean.
- Removed `.github/dependabot.yml`: Renovate (configured 2026-06-14, manages gomod +
  github-actions) fully superseded it, so both bots were racing to open duplicate
  dependency PRs. Renovate is now the sole dependency manager.
