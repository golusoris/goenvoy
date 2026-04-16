# AGENTS — arr/prowlarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://prowlarr.com/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-prowlarr.md](../../docs/upstream/arr-prowlarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

History and log endpoints take `page` and `pageSize` query params; indexer history takes `limit`.
