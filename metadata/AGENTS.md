# AGENTS — metadata

> Per-category conventions for **metadata/** modules. Read repo-root [AGENTS.md](../AGENTS.md) + [.workingdir/PRINCIPLES.md](../.workingdir/PRINCIPLES.md) first.

## What this category is

Clients for external metadata providers: movies, TV, anime, music, games, books, tracking services, and adult content. Consumed by media managers, enrichment pipelines, and custom automation. Grouped by media type so that a caller interested in "movie posters" doesn't have to page through "game cover art".

## Shared types (in `metadata/`)

The category root ([metadata/types.go](types.go)) defines types shared across sub-categories — `Rating`, `Image`, `Person`, `Episode`, `Season`. Service sub-modules (e.g. `metadata/video/tmdb`) **may** import `metadata` for these types but not each other.

## Sub-categories

| Sub-category | Services | Typical auth |
|---|---|---|
| [video/](video/) | TMDB, TVDB, TVmaze, Fanart.tv, OMDb, Letterboxd, OpenSubtitles | API key (header or query) |
| [anime/](anime/) | AniList, MAL, AniDB, Kitsu | API key / OAuth / GraphQL-bearer |
| [music/](music/) | MusicBrainz, Discogs, Spotify, Deezer, Last.fm, AudioDB, ListenBrainz | API key / OAuth / none |
| [tracking/](tracking/) | Trakt, SIMKL | OAuth 2.0 (device code / PKCE) |
| [book/](book/) | Open Library, Google Books | API key / none |
| [game/](game/) | IGDB, RAWG, Steam, SteamGridDB, Hasheous, LaunchBox, MobyGames, RetroAchievements, ScreenScraper | API key / OAuth (IGDB via Twitch) |
| [adult/](adult/) | ThePornDB, StashDB | Bearer token |

## Shared conventions in this category

- **Auth diversity**: no single auth pattern. Each service documents its own in its `AGENTS.md`.
- **Base-URL shape**: full upstream base, service-specific. Most use `https://api.<provider>.tld/v<n>`.
- **Pagination**: varies wildly — page+pageSize, cursor, offset+limit, Link header. Document per service.
- **Error body**: JSON envelope per provider — TMDB/TVDB have well-defined shapes; others vary.
- **Rate limits**: non-negligible for most — see each service's `AGENTS.md`. Clients currently do **not** implement built-in retry/backoff; callers must add their own `http.Client` / middleware.
- **Date format**: RFC 3339 preferred; some providers use `YYYY-MM-DD` only. Parse defensively.

## Modules in this category

34 service modules total. See each sub-category's directory (links above) for the full list.

## When adding a new service here

Use `/add-service-client metadata/<sub-category> <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../_meta/service-agents-template.md)).

If the new service belongs to a sub-category that already exists, drop it there. If it's genuinely new (e.g. "podcast metadata"), create a new sub-category directory and open a PR introducing it — mention the sub-category choice in the PR description for review.
