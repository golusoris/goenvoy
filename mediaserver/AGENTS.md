# AGENTS — mediaserver

> Per-category conventions for **mediaserver/** modules. Read repo-root [AGENTS.md](../AGENTS.md) + [.workingdir/PRINCIPLES.md](../.workingdir/PRINCIPLES.md) first.

## What this category is

Clients for self-hosted media-server applications: Plex, Jellyfin, Emby, Kavita, Komga, Navidrome, Audiobookshelf, Tautulli, Stash, Tdarr. Covers both "library/metadata reading" and "playback/session management" where exposed.

## Shared types (in `mediaserver/`)

The category root ([mediaserver/types.go](types.go)) defines shared cross-service types — `Library`, `Item`, `User`. These are intentionally **loose** because the upstream mediaserver APIs diverge heavily.

Service sub-modules **may** import `mediaserver` for the shared types but not each other.

## Shared conventions in this category

- **Auth diversity**: no single pattern.
  - Plex — `X-Plex-Token` header (legacy; OAuth also supported on plex.tv).
  - Jellyfin / Emby — `X-Emby-Token` header (same header name for both, historically).
  - Navidrome — Subsonic API (`u`, `t`, `s` query params) or native token.
  - Audiobookshelf — `Authorization: Bearer <token>`.
  - Kavita / Komga — Bearer JWT, rotated via refresh endpoint.
  - Tautulli — `?apikey=` query.
  - Stash — GraphQL over HTTP POST with `ApiKey` header.
  - Tdarr — Bearer token.
- **Base-URL shape**: callers pass the mediaserver's root (`http(s)://<host>[:<port>]`); clients append the API prefix (`/web/api/...`, `/api/...`, etc.) internally.
- **Pagination**: service-specific. Plex/Emby/Jellyfin use query params (`startIndex`, `limit`); Stash uses GraphQL pagination; others vary. Document per service.
- **Error body**: JSON per service. Stash returns GraphQL `errors[]`; the rest return `{error: string}` or a `Message` field on 4xx.
- **Content negotiation**: some services (Plex especially) distinguish JSON vs XML via `Accept:` header — default to JSON everywhere.

## Modules in this category

| Module | Purpose | Upstream pin |
|---|---|---|
| [mediaserver/plex](plex/) | Plex Media Server | [docs/upstream/mediaserver-plex.md](../docs/upstream/mediaserver-plex.md) |
| [mediaserver/jellyfin](jellyfin/) | Jellyfin | [docs/upstream/mediaserver-jellyfin.md](../docs/upstream/mediaserver-jellyfin.md) |
| [mediaserver/emby](emby/) | Emby | [docs/upstream/mediaserver-emby.md](../docs/upstream/mediaserver-emby.md) |
| [mediaserver/kavita](kavita/) | Kavita (comics/manga/books) | [docs/upstream/mediaserver-kavita.md](../docs/upstream/mediaserver-kavita.md) |
| [mediaserver/komga](komga/) | Komga (comics) | [docs/upstream/mediaserver-komga.md](../docs/upstream/mediaserver-komga.md) |
| [mediaserver/navidrome](navidrome/) | Navidrome (Subsonic API) | [docs/upstream/mediaserver-navidrome.md](../docs/upstream/mediaserver-navidrome.md) |
| [mediaserver/audiobookshelf](audiobookshelf/) | Audiobookshelf | [docs/upstream/mediaserver-audiobookshelf.md](../docs/upstream/mediaserver-audiobookshelf.md) |
| [mediaserver/tautulli](tautulli/) | Tautulli (Plex stats) | [docs/upstream/mediaserver-tautulli.md](../docs/upstream/mediaserver-tautulli.md) |
| [mediaserver/stash](stash/) | Stash (GraphQL) | [docs/upstream/mediaserver-stash.md](../docs/upstream/mediaserver-stash.md) |
| [mediaserver/tdarr](tdarr/) | Tdarr (transcode server) | [docs/upstream/mediaserver-tdarr.md](../docs/upstream/mediaserver-tdarr.md) |

## When adding a new service here

Use `/add-service-client mediaserver <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../_meta/service-agents-template.md)).

Mediaserver auth is the most frequent source of surprise — document the exact token-placement and any refresh behaviour in the service's `AGENTS.md` before wiring methods.
