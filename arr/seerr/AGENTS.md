# AGENTS — arr/seerr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api-docs.overseerr.dev/>
- Last verified: 2026-10-16 (sct/overseerr@98ea135f2d370faba0362d259fccf51d8fd44a16 https://raw.githubusercontent.com/sct/overseerr/develop/overseerr-api.yml, seerr-team/seerr@978f437dc06af77459c8887781a32d2879f1b69d https://raw.githubusercontent.com/seerr-team/seerr/develop/seerr-api.yml)
- Details: [docs/upstream/arr-seerr.md](../../docs/upstream/arr-seerr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

Search and discover endpoints take a `page` query param.

## Capabilities

- MediaRequest Type and Seasons (*[]RequestSeason; nil means response carried no seasons field, e.g. movie requests) decoding for TV show requests.
- MediaInfo MediaType and Status4k decoding for media classification and 4K availability.
- User DisplayName decoding and (*User).Name() fallback resolution.
- ServiceSettings retrieval for Sonarr (GetSonarrSettings) and Radarr (GetRadarrSettings).
- ServiceSettings.URL() helper for rebuilding downstream service URLs with proper IPv6 and path handling.

## Known quirks

- Client targets Overseerr; Jellyseerr is wire-compatible with the same endpoints.
