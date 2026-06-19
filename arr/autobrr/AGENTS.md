# AGENTS — arr/autobrr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://autobrr.com/api>
- Last verified: 2026-06-14
- Details: [docs/upstream/arr-autobrr.md](../../docs/upstream/arr-autobrr.md)

## Auth model

API key passed as `X-Api-Token` header.

## Pagination

The release list endpoint takes `offset` and `limit` query params.
