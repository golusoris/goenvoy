---
name: release-module
description: Tag a goenvoy module by hand and trigger the release workflow, outside the release pull request.
---

# Skill — `/release-module`

Cut the tag that triggers [release.yml](../../.github/workflows/release.yml) (cosign + SBOM + SLSA provenance + GH Release).

## When to use

Regular releases do not use this skill: [release-please.yml](../../.github/workflows/release-please.yml) keeps one release pull request current, and merging it creates the tags and the signed releases for every module it lists. To retry a release whose run failed, dispatch `release.yml` with the existing tag.

Use this skill only when the user asks for a hand-cut tag of a version the release manifest (`.release-please-manifest.json`) already names.

## Expected arguments

- `$1` — module path, e.g. `arr/sonarr`.
- `$2` — version, e.g. `v1.4.0`. Must match `^v\d+\.\d+\.\d+(-[A-Za-z0-9.-]+)?$`.

## Steps

1. Verify current branch is `main` and working tree is clean. If not, abort.
2. `git pull --ff-only origin main`.
3. Verify `<module>/CHANGELOG.md` has the release-please stanza for `<version>`; else abort (the release pull request did not ship it).
4. Run the pre-release audit: `tools/release-check.sh <module> <version>`. Must pass.
5. Verify the tag doesn't already exist: `git tag --list "<module>/<version>"` must be empty. Also check remote: `git ls-remote --tags origin "<module>/<version>"`.
6. Create and push the tag:
   ```bash
   git tag "<module>/<version>"
   git push origin "<module>/<version>"
   ```
7. Wait briefly (`sleep 10`) then report the release.yml run URL: `gh run list --workflow=release.yml --limit=1 --json url,status,conclusion`.
8. Tell the user to watch the run. When it completes, a signed GH Release with the tarball, checksums, .sig, .pem, and SBOM will be published at `https://github.com/golusoris/goenvoy/releases/tag/<module>/<version>`.
9. Remind the user: to consume the new version from another module in this monorepo, bump the `require` line in that module's `go.mod` and run `go mod tidy`.

## Don't

- Don't tag `v0.0.0` — semver starts at `v0.1.0` or `v1.0.0`.
- Don't tag from a dirty tree. A stray uncommitted file can end up in the source tarball.
- Don't release without the `<module>/CHANGELOG.md` stanza — the GH Release notes link to it.
- Don't retry by force-pushing a tag. If the release workflow failed, investigate, fix forward, and cut a new patch version. Moving tags is forbidden.
