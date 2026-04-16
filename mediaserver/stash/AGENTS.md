# AGENTS — mediaserver/stash

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://docs.stashapp.cc/in-app-manual/>
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-stash.md](../../docs/upstream/mediaserver-stash.md)

## Auth model

Optional API key passed as `Apikey` header against the GraphQL endpoint.

## Pagination

GraphQL filter input supports `per_page` and a page cursor.
