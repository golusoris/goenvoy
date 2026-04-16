# AGENTS — mediaserver/emby

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://dev.emby.media/doc/restapi/>
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-emby.md](../../docs/upstream/mediaserver-emby.md)

## Auth model

Access token passed as `X-Emby-Token` header; obtained via `AuthenticateByName` or `WithAccessToken`.

## Pagination

`GetItemsByParent` and related endpoints take `StartIndex` and `Limit` query params.
