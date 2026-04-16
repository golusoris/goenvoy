# AGENTS — arr/seerr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api-docs.overseerr.dev/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-seerr.md](../../docs/upstream/arr-seerr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

Search and discover endpoints take a `page` query param.

## Known quirks

- Client targets Overseerr; Jellyseerr is wire-compatible with the same endpoints.
