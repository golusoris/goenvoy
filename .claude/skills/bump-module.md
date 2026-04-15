---
name: bump-module
description: Bump one goenvoy module's semver and open a release PR.
---

# Skill — `/bump-module`

Bump a single module's version and open the PR that will ship it.

## When to use

The user says "bump <module> <level>" — e.g. "bump arr/sonarr minor". The module already has a `## Unreleased` section in the root `CHANGELOG.md` that you're about to promote.

## Expected arguments

- `$1` — module path relative to repo root, e.g. `arr/sonarr`, `metadata/video/tmdb`.
- `$2` — bump level: `major | minor | patch`.

## Steps

1. Find the latest tag for this module: `git tag --list "<module>/v*" --sort=-v:refname | head -1`.
2. Compute the next version from `$2`:
   - `major` — `vN.0.0` (next major; reset minor+patch; only if an ADR justifies it and there's a `BREAKING CHANGE:` footer in the pending commits).
   - `minor` — `vN.M+1.0` (new feature, backwards compatible).
   - `patch` — `vN.M.P+1` (fix or dependency-free chore).
3. Verify the contents of the module's `## Unreleased` stanza in [CHANGELOG.md](../../CHANGELOG.md) are non-empty. If empty, abort: nothing to release.
4. In `CHANGELOG.md`, promote `## Unreleased` → `## <module>/<new-version> — YYYY-MM-DD` (today's date). Add a fresh empty `## Unreleased` stub above it.
5. Run the pre-release audit: `tools/release-check.sh <module> <new-version>`. If it fails (breaking API change without a major bump, tests failing, lint failing) stop and report.
6. Create a branch `release/<module>-<new-version-without-leading-v>`, commit the CHANGELOG change with message `release(<module>): <new-version>`, push.
7. Open a PR with `gh pr create`:
   - Title: `release(<module>): <new-version>`.
   - Body: the CHANGELOG stanza you just promoted, plus a checkbox list: `- [ ] CHANGELOG entry reviewed`, `- [ ] apidiff clean (see CI)`, `- [ ] Ready to tag after merge`.
8. Report the PR URL to the user. After the PR is merged, run `/release-module <module> <new-version>` to cut the tag.

## Don't

- Don't bump major without a `BREAKING CHANGE:` footer in at least one commit since the prior tag. The CI apidiff gate will fail anyway.
- Don't edit an already-released version stanza in `CHANGELOG.md`.
- Don't tag from this skill — tagging happens in `/release-module`, post-merge.
