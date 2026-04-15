# AGENTS — arr

> Per-category conventions for **arr/** modules. Read repo-root [AGENTS.md](../AGENTS.md) + [.workingdir/PRINCIPLES.md](../.workingdir/PRINCIPLES.md) first.

## What this category is

Clients for the *arr stack and adjacent automation tools — Sonarr, Radarr, Readarr, Lidarr, Bazarr, Prowlarr, Whisparr, Mylar, Jackett, NZBHydra2, Autobrr, FlareSolverr, Overseerr/Jellyseerr (`seerr`).

## Shared types (in `arr/`)

The root of the category is itself a small module ([arr/types.go](types.go)) containing types shared across services — `SystemStatus`, `RootFolder`, `Tag`, `QualityProfile`, `LanguageProfile`. Every *arr service returns these in (near-)identical shape.

Service sub-modules **may** import `arr` for the shared types. This is the only in-repo import edge allowed — no service imports another service.

## Shared conventions in this category

- **Auth model**: `X-Api-Key` request header on every call. Some services (Jackett, NZBHydra2) accept `?apikey=<key>` query instead — note per-service in the service `AGENTS.md`.
- **Base-URL shape**: callers pass `http(s)://<host>[:<port>]`. The client appends `/api/v<major>` (usually `v3` for Sonarr/Radarr/Prowlarr, `v1` for Readarr/Lidarr, service-specific for the rest).
- **Pagination**: none — endpoints return complete lists. A few "event"-style endpoints (`/history`, `/queue`) take `page` / `pageSize` query params.
- **Error body**: JSON `{"error": "...", "propertyName": "..."}` on 4xx; plain text / HTML on 5xx.
- **Date format**: RFC 3339 (`2026-04-15T12:34:56Z`).

## Modules in this category

| Module | Purpose | Upstream pin |
|---|---|---|
| [arr/sonarr](sonarr/) | Sonarr v3 client | [docs/upstream/arr-sonarr.md](../docs/upstream/arr-sonarr.md) |
| [arr/radarr](radarr/) | Radarr v3 client | [docs/upstream/arr-radarr.md](../docs/upstream/arr-radarr.md) |
| [arr/readarr](readarr/) | Readarr v1 client | [docs/upstream/arr-readarr.md](../docs/upstream/arr-readarr.md) |
| [arr/lidarr](lidarr/) | Lidarr v1 client | [docs/upstream/arr-lidarr.md](../docs/upstream/arr-lidarr.md) |
| [arr/whisparr](whisparr/) | Whisparr v3 client | [docs/upstream/arr-whisparr.md](../docs/upstream/arr-whisparr.md) |
| [arr/bazarr](bazarr/) | Bazarr subtitle client | [docs/upstream/arr-bazarr.md](../docs/upstream/arr-bazarr.md) |
| [arr/prowlarr](prowlarr/) | Prowlarr indexer client | [docs/upstream/arr-prowlarr.md](../docs/upstream/arr-prowlarr.md) |
| [arr/jackett](jackett/) | Jackett indexer client | [docs/upstream/arr-jackett.md](../docs/upstream/arr-jackett.md) |
| [arr/nzbhydra](nzbhydra/) | NZBHydra2 meta-indexer | [docs/upstream/arr-nzbhydra.md](../docs/upstream/arr-nzbhydra.md) |
| [arr/mylar](mylar/) | Mylar3 comics client | [docs/upstream/arr-mylar.md](../docs/upstream/arr-mylar.md) |
| [arr/autobrr](autobrr/) | Autobrr filter-rules client | [docs/upstream/arr-autobrr.md](../docs/upstream/arr-autobrr.md) |
| [arr/flaresolverr](flaresolverr/) | FlareSolverr CF-bypass client | [docs/upstream/arr-flaresolverr.md](../docs/upstream/arr-flaresolverr.md) |
| [arr/seerr](seerr/) | Overseerr / Jellyseerr client | [docs/upstream/arr-seerr.md](../docs/upstream/arr-seerr.md) |

## When adding a new service here

Use `/add-service-client arr <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../_meta/service-agents-template.md)).
