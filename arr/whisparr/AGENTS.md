# AGENTS — arr/whisparr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://wiki.servarr.com/whisparr>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-whisparr.md](../../docs/upstream/arr-whisparr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

Paged endpoints (queue, history, wanted, blocklist, log) take `page` and `pageSize` query params.

## API notes

Queue management: `GetQueueWithOptions` and `DeleteQueueItemWithOptions` support advanced filtering, nested inclusion, and extended deletion flags across v2 (Sonarr-based) and v3 (Radarr-based) clients.

## Known quirks

- Ships V2 (Eros) and V3 client variants — V2 endpoints live in `eros.go`.
