# AGENTS — arr/readarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://readarr.com/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-readarr.md](../../docs/upstream/arr-readarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

Paged endpoints (queue, history, wanted, blocklist, log) take `page` and `pageSize` query params.
