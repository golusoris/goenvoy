# Session state — goenvoy

> Persistent state across workstations and AI sessions. Updated as significant changes happen.
> Last update: 2026-09-02 (Gitea publication and Discogs test isolation).

## 2026-09-02 Notes

- Restricted the Gitea-to-GitHub `workflow_run` publisher to successful `main`
  CI events. Pull-request completions had entered the same publication
  concurrency group and cancelled the waiting validated-main publisher.
- Isolated the Discogs tests from `http.DefaultTransport` by using each
  `httptest.Server` client. The shared transport let one parallel test's server
  cleanup interrupt another test's POST with `http: CloseIdleConnections
  called`; this caused the 2026-09-02 Gitea-main CI failure.
- Reproduced the documented `make ci-all` gate with Go 1.26.7,
  golangci-lint 2.13.2, gosec 2.27.1, and govulncheck 1.4.0 under the office
  runner's 4-CPU/6-GiB envelope. Every module passed lint, vet, security,
  race/coverage, and build in 367 seconds; peak memory was 1,341,501,440 bytes
  with no cgroup pressure or OOM events.
- Disabled platform-native Renovate auto-merge for low-risk bumps. Gitea has no
  required-status branch gate here, so native auto-merge could land work before
  CI started or while it was failing. Renovate's own merge path waits for green,
  requires the branch to be current, and merges at most one PR per target branch
  per run.
- Keyed publication concurrency by event and ref. Validated `main` events still
  serialize, but Goenvoy's distinct module tags can no longer cancel one another
  while multiple release publications wait for the office runner.
- Live acceptance remains an exact Gitea-main CI success followed by publication
  of that same SHA to public GitHub.

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
