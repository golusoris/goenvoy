# AGENTS — anime/shoko

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://docs.shokoanime.com/>
- Last verified: 2026-04-15
- Details: [docs/upstream/anime-shoko.md](../../docs/upstream/anime-shoko.md)

## Auth model

`Login(username, password)` POSTs to `/api/auth`; the returned API key is sent as `Apikey` header on subsequent requests.

## Pagination

List endpoints take `page` and `pageSize` query params.
