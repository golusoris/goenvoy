---
name: bump-module
description: Bump one goenvoy module's semver and open a release PR.
---

# Skill — `/bump-module`

Bump a single module's version and open the PR that will ship it.

## When to use

The user says "bump <module> <level>" — e.g. "bump arr/sonarr minor". release-please computes versions from Conventional Commits; this skill only forces a version, through the per-package `release-as` key in `release-please-config.json`.

## Expected arguments

- `$1` — module path relative to repo root, e.g. `arr/sonarr`, `metadata/video/tmdb`.
- `$2` — bump level: `major | minor | patch`.

## Steps

1. Find the latest tag for this module: `git tag --list "<module>/v*" --sort=-v:refname | head -1`.
2. Compute the next version from `$2`:
   - `major` — `vN.0.0` (next major; reset minor+patch; only if an ADR justifies it and there's a `BREAKING CHANGE:` footer in the pending commits).
   - `minor` — `vN.M+1.0` (new feature, backwards compatible).
   - `patch` — `vN.M.P+1` (fix or dependency-free chore).
3. Run the pre-release audit: `tools/release-check.sh <module> <new-version>`. If it fails (breaking API change without a major bump, tests failing, lint failing) stop and report.
4. Create a branch `release/<module>-<new-version-without-leading-v>`. In `release-please-config.json`, add `"release-as": "<new-version-without-leading-v>"` to the module's entry under `packages`. Commit with message `chore(<module>): release <new-version>`, push.
5. Open a PR with `gh pr create`:
   - Title: `chore(<module>): release <new-version>`.
   - Body: the forced version and why, plus a checkbox list: `- [ ] apidiff clean (see CI)`, `- [ ] Remove release-as after the release ships`.
6. Report the PR URL to the user. After it merges, the release pull request carries the version; merging that one creates the tag. Then remove the `release-as` key again, or the next release repeats the version.

## Don't

- Don't bump major without a `BREAKING CHANGE:` footer in at least one commit since the prior tag. The CI apidiff gate will fail anyway.
- Don't edit any `CHANGELOG.md`; release-please writes it.
- Don't tag from this skill — merging the release pull request tags.
