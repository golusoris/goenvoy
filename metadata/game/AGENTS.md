# AGENTS — metadata/game

> Sub-category root for **metadata/game/** modules. Read [../AGENTS.md](../AGENTS.md) + repo-root [AGENTS.md](../../AGENTS.md) first.

## What this sub-category is

Game-metadata providers — game catalogs, ratings, artwork, achievements, ROM hashes, retro-game databases.

## Shared types (in `metadata/game/`)

This directory is **itself a Go module** (`github.com/golusoris/goenvoy/metadata/game`). It defines types shared across the game-provider service modules — see [types.go](types.go). Service sub-modules (e.g. `metadata/game/igdb`) **may** import `metadata/game` for these types but not each other.

## Shared conventions in this sub-category

- **Auth diversity**: largely API key (RAWG, MobyGames, SteamGridDB, ScreenScraper) or Twitch OAuth2 (IGDB). Each service documents its own.
- **Base-URL shape**: full upstream base, service-specific.
- **Pagination**: varies — page+pageSize (RAWG, MobyGames), offset+limit (IGDB), cursorless full-list (Hasheous).
- **Error body**: provider-specific JSON envelope. Always surfaced as `*APIError`.
- **Image URLs**: every provider returns CDN URLs in different shapes; `metadata/game.Image` provides a normalised representation. Conversion happens in each service's `types.go`.

## Modules in this sub-category

| Module | Purpose | Upstream pin |
|---|---|---|
| [igdb/](igdb/) | IGDB (Twitch-OAuth) | [docs/upstream/metadata-game-igdb.md](../../docs/upstream/metadata-game-igdb.md) |
| [rawg/](rawg/) | RAWG | [docs/upstream/metadata-game-rawg.md](../../docs/upstream/metadata-game-rawg.md) |
| [steam/](steam/) | Steam Web API | [docs/upstream/metadata-game-steam.md](../../docs/upstream/metadata-game-steam.md) |
| [steamgriddb/](steamgriddb/) | SteamGridDB | [docs/upstream/metadata-game-steamgriddb.md](../../docs/upstream/metadata-game-steamgriddb.md) |
| [hasheous/](hasheous/) | Hasheous (ROM-hash lookup) | [docs/upstream/metadata-game-hasheous.md](../../docs/upstream/metadata-game-hasheous.md) |
| [launchbox/](launchbox/) | LaunchBox Games Database | [docs/upstream/metadata-game-launchbox.md](../../docs/upstream/metadata-game-launchbox.md) |
| [mobygames/](mobygames/) | MobyGames | [docs/upstream/metadata-game-mobygames.md](../../docs/upstream/metadata-game-mobygames.md) |
| [retroachievements/](retroachievements/) | RetroAchievements | [docs/upstream/metadata-game-retroachievements.md](../../docs/upstream/metadata-game-retroachievements.md) |
| [screenscraper/](screenscraper/) | ScreenScraper | [docs/upstream/metadata-game-screenscraper.md](../../docs/upstream/metadata-game-screenscraper.md) |

## When adding a new service here

Use `/add-service-client metadata/game <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`. Map upstream shapes to the shared types in `metadata/game/types.go` where meaningful.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../../_meta/service-agents-template.md)).
