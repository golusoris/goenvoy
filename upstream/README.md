# Upstream OpenAPI spec pins

`specs.yaml` pins the OpenAPI spec of each upstream project once, for goenvoy and for every
project that checks a client against those specs. An entry names the upstream repository, the
release tag (or, for a project that tags no release, the branch), the commit, the path of the spec
file at that commit, its sha256, the SPDX licence of the upstream repository, and the request path
prefix the spec leaves to its `servers` entry.

## No spec bytes

Only the pin is committed, never the spec. The Sonarr, Radarr, Lidarr and Readarr-fork specs are
GPL-3.0 and this repository is MIT, so neither the spec files nor anything built from them enters
this repository or a module release. The checker downloads each file, hashes it in memory and
writes nothing to disk.

## Check

```bash
make upstream-specs
```

runs the checker's tests, which refuse planted defects (a malformed entry, another sha256, a
commit or path that does not serve the file, an oversized reply, a Renovate configuration that
does not read every pin), and then checks every pin against the live file. The CI job
`Upstream spec pins` runs the same target on every pull request.

## Renewing a pin

Renovate reads repository, tag and commit of each entry (the two `customManagers` for
`upstream/specs.yaml` in `renovate.json`) and proposes the next stable release, or the newest
commit of a pinned branch. When the spec file is unchanged at the new commit, the sha256 still
fits and the check passes. When it changed, the check fails and prints the sha256 it found; a
person reviews the spec change, puts the new sha256 into the entry, and merges. A pin bump is
never merged automatically, because the sha256 shows only whether the file changed, not whether
each consumer's client still fits.

## Consumers

A consumer reads this file at a pinned goenvoy commit, with the sha256 of `specs.yaml` itself,
and keeps its own list of the operations it checks. lusoris/SnatchArr does this for its Go and Rust
contract tests (`tools/upstreamspec` there).
