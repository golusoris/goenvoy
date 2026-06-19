# Session state — goenvoy

> Persistent state across workstations and AI sessions. Updated as significant changes happen.
> Last update: 2026-06-19 (verified the hardening pass, bumped govulncheck).

## Standards rollout — `.goenvoy2.0/`

Phased adoption of golusoris standards. Plan in [`.goenvoy2.0/`](../.goenvoy2.0/), checklist in [`.goenvoy2.0/10-rollout-checklist.md`](../.goenvoy2.0/10-rollout-checklist.md).

| Phase | Title | Status |
|---|---|---|
| 1 | Governance docs + editor config | Done (commit `0e4352b`) |
| 2 | Principles + ADRs + working dir | In progress |
| 3 | Lint + tooling baseline | In progress |
| 4 | CI hardening (CodeQL, Scorecard, apidiff) | In progress |
| 5 | Release + supply-chain hardening | In progress |
| 6 | Claude hooks + skills + per-module AGENTS | In progress |

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
- GitHub state: Renovate PR #69 bumps `actions/checkout` v6 → v7. `codeql.yml` pins
  an orphaned older v6 checkout SHA (`de0fac2e`) that PR #69 does not update. The
  Renovate Dependency Dashboard (#62) lists no unopened version bumps beyond #69.
  release-please PR #68 is pending (releases the modules touched by the arr/v2 →
  v2.1.0 dependency bumps).
