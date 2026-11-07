<!-- markdownlint-disable MD013 MD023 MD033 MD036 MD041 -->
<!-- A centred header needs inline HTML, a banner image, indented headings and
     badge lines past 80 columns, so those five rules are off for this file.
     Same directive style as docs/adr/0000-template.md. The prose below still
     wraps at 80. -->
<div align="center">

  <img src="docs/assets/goenvoy-readme-banner.svg" alt="goenvoy — pure-stdlib Go client libraries for media automation, metadata, and download APIs" height="120" />

  <!-- README-STANDARD badge block: one group, fixed order (docs/README-STANDARD.md) -->
  <!-- 1 release/version -->
  [![Release](https://img.shields.io/badge/Release-Releases-3b82f6?style=for-the-badge&labelColor=111a35)](https://github.com/golusoris/goenvoy/releases)
  [![Go](https://img.shields.io/badge/Go-1.26+-06b6d4?style=for-the-badge&logo=go&logoColor=white&labelColor=111a35)](https://go.dev)
  <!-- 2 CI/build -->
  [![CI](https://github.com/golusoris/goenvoy/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/golusoris/goenvoy/actions/workflows/ci.yml)
  <!-- 3 security/scorecard -->
  [![Security](https://github.com/golusoris/goenvoy/actions/workflows/security-scan.yml/badge.svg?branch=main&event=push)](https://github.com/golusoris/goenvoy/actions/workflows/security-scan.yml)
  [![OpenSSF Scorecard](https://img.shields.io/badge/OpenSSF-Scorecard-06b6d4?style=for-the-badge&labelColor=111a35)](https://scorecard.dev/viewer/?uri=github.com/golusoris/goenvoy)
  [![OpenSSF Best Practices](https://www.bestpractices.dev/projects/15901/badge)](https://www.bestpractices.dev/projects/15901)
  [![Codacy](https://app.codacy.com/project/badge/Grade/40988f0fd89c4cd5a17c857fe5bac649)](https://app.codacy.com/gh/golusoris/goenvoy/dashboard)
  [![CodeRabbit](https://img.shields.io/badge/Reviews-CodeRabbit-8b5cf6?style=for-the-badge&logo=coderabbit&logoColor=white&labelColor=111a35)](https://coderabbit.ai)
  <!-- 4 standards -->
  [![HISS-21](https://img.shields.io/badge/Standards-HISS--21-06b6d4?style=for-the-badge&logo=nasa&labelColor=111a35)](AGENTS.md)
  [![Power of 10](https://img.shields.io/badge/NASA_JPL-Power_of_10-3b82f6?style=for-the-badge&logo=nasa&labelColor=111a35)](docs/principles.md)
  [![Conventional Commits](https://img.shields.io/badge/Commits-Conventional-8b5cf6?style=for-the-badge&logo=conventionalcommits&logoColor=white&labelColor=111a35)](CONTRIBUTING.md)
  <!-- 5 licence -->
  [![License](https://img.shields.io/badge/License-MIT-3b82f6?style=for-the-badge&labelColor=111a35)](LICENSE)
  <!-- 6 support -->
  [![GitHub Sponsors](https://img.shields.io/badge/Sponsor-GitHub-8b5cf6?style=for-the-badge&logo=githubsponsors&logoColor=white&labelColor=111a35)](https://github.com/sponsors/lusoris)
  [![Ko-fi](https://img.shields.io/badge/Support-Ko--fi-8b5cf6?style=for-the-badge&logo=kofi&logoColor=white&labelColor=111a35)](https://ko-fi.com/lusoris)

  **Pure-stdlib Go client libraries for media automation, metadata, and download APIs**

  *69 modules covering 63+ services in a multi-module monorepo with zero external dependencies*

</div>

goenvoy is a collection of Go client libraries for media automation, metadata,
and download service APIs — 69 modules covering 63+ services. Designed as a
multi-module monorepo, it allows importing only what you need, from a single
library to an entire category, using only Go's standard library.

**[📖 Principles](docs/principles.md)** · [Architecture Decisions](docs/adr/) · [Contributing](CONTRIBUTING.md)

---

## Why goenvoy

- **Pure stdlib** — zero external dependencies, just `net/http` and `encoding/json`
- **Context-aware** — every method takes `context.Context` for cancellation and timeouts
- **Functional options** — `WithHTTPClient`, `WithTimeout`, `WithHeader` and service-specific options
- **Type-safe errors** — each module defines an `APIError` with status code, message, and body
- **OAuth2 support** — device code, authorization code, PKCE, and refresh flows where applicable
- **Thoroughly tested** — httptest-based tests, race detector enabled, no live API calls

## Get started

### Install

Import a specific service library:

```go
go get github.com/golusoris/goenvoy/arr/sonarr
```

Or import shared category types:

```go
go get github.com/golusoris/goenvoy/arr
```

### Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/golusoris/goenvoy/arr/sonarr"
)

func main() {
    client, err := sonarr.New("http://localhost:8989", "your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get system status
    status, err := client.GetSystemStatus(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s version %s\n", status.AppName, status.Version)

    // Get all series
    series, err := client.GetAllSeries(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Total series: %d\n", len(series))
}
```

Each service module follows the same pattern: `New(baseURL, apiKey) → typed methods with context`.

## At a glance

### Categories

| Category | Module | Services |
| --- | --- | --- |
| **Arr Stack** | `github.com/golusoris/goenvoy/arr` | Sonarr, Radarr, Lidarr, Readarr, Whisparr (v2+v3), Prowlarr, Bazarr, Seerr, Autobrr, Mylar3, Jackett, NZBHydra2 |
| **Metadata** | `github.com/golusoris/goenvoy/metadata` | TMDb, TheTVDB, Fanart.tv, OMDb, TVmaze, AniList, Kitsu, AniDB, MAL, Trakt, Simkl, Letterboxd, ListenBrainz, MusicBrainz, StashBox, TPDB, OpenSubtitles, Last.fm, Discogs, TheAudioDB, Open Library, Google Books, Spotify, Deezer, IGDB, RAWG, Steam, MobyGames, SteamGridDB, RetroAchievements, ScreenScraper, Hasheous, LaunchBox |
| **Download Clients** | `github.com/golusoris/goenvoy/downloadclient` | qBittorrent, Transmission, Deluge, rTorrent, SABnzbd, NZBGet |
| **Media Servers** | `github.com/golusoris/goenvoy/mediaserver` | Plex, Jellyfin, Emby, Tautulli, Audiobookshelf, Komga, Navidrome, Kavita, Stash, Tdarr |
| **Tools** | `github.com/golusoris/goenvoy/tools` | FlareSolverr |
| **Anime** | `github.com/golusoris/goenvoy/anime` | Shoko Server |

### Structure

Each category has a base module with shared types, plus sub-modules for individual services:

```text
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
│   ├── cleanuparr/   # Cleanuparr client
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
│   ├── nzbget/       # NZBGet JSON-RPC client
│   └── queue/        # Unified read-only queue snapshot
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
├── anime/            # shared anime types
│   └── shoko/        # Shoko Server client
```

## Contribute

Requires Go 1.26+. See [CONTRIBUTING.md](CONTRIBUTING.md) for full details.

```bash
# Set up workspace (local dev, links all 63 modules)
go work init && find . -name 'go.mod' -not -path './.workingdir/*' -exec dirname {} \; | xargs go work use

# Run all tests
make test-all

# Lint all modules
make lint-all

# Tidy all modules
make tidy-all

# Format all modules
make fmt-all
```

## Support

If goenvoy saves you time, a coffee helps ☕

<p align="left">
  <a href="https://ko-fi.com/lusoris" target="_blank">
    <img src="https://ko-fi.com/img/githubbutton_sm.svg" alt="Support me on Ko-fi" />
  </a>
  &nbsp;&nbsp;
  <a href="https://github.com/sponsors/lusoris" target="_blank">
    <img src="https://img.shields.io/badge/Sponsor-%E2%9D%A4-ea4aaa?style=flat-square&logo=github&logoColor=white" alt="GitHub Sponsors" />
  </a>
</p>

## Upstream and licence

[MIT](LICENSE)

## Standards & Governance

- Pure stdlib (ADR-0001). `depguard` in CI enforces.
- Per-module independent semver (ADR-0002). Tags: `<path>/vX.Y.Z`.
- Every release: cosign-signed checksums + SPDX SBOM + SLSA-L3 provenance.
- Every merged commit: 0 lint / 0 gosec / 0 govulncheck / race-green across all modules.
- [AGENTS.md](AGENTS.md) / [CLAUDE.md](CLAUDE.md) for agent-assisted development conventions.
- `docs/adr/` for all architectural decisions (Nygard format).

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

<!-- praetor:readme-governance:start -->
[![Documentation Governance][praetor-docs-badge]][praetor-docs-runs]

Praetor manages this repository's declared governance policy. This managed
block records adoption state; it is not a verification certificate.

**Verification**: `make verify-all` runs the repository's configured
verification cascade.

**HISS Audit**: `praetorctl audit` enforces policy, generated-surface
integrity, and the debt ratchet.

**Context Sync**: `praetorctl compile-context --verify` verifies every
generated agent context against `AGENTS.md`.

**Documentation**: `make docs-lint` enforces locked Markdown style and the
private scratch-link policy.

**Debt Baseline**: `.standards-baseline.json` anchors the debt ratchet at
5 recorded infractions; audit forbids growth.

[praetor-docs-badge]: https://github.com/golusoris/goenvoy/actions/workflows/praetor-docs.yml/badge.svg
[praetor-docs-runs]: https://github.com/golusoris/goenvoy/actions/workflows/praetor-docs.yml
<!-- praetor:readme-governance:end -->
