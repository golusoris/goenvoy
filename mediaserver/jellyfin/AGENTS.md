# AGENTS — mediaserver/jellyfin

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api.jellyfin.org/>
- Last verified: 2026-10-23
- Details: [docs/upstream/mediaserver-jellyfin.md](../../docs/upstream/mediaserver-jellyfin.md)

## Auth model

`Authorization: MediaBrowser Client=..., Device=..., DeviceId=..., Version=..., Token=...` header; token is obtained via `AuthenticateByName`, `AuthenticateUserByName`, `AuthenticateWithQuickConnect` (after `InitiateQuickConnect` + `AuthorizeQuickConnect`) or `WithAccessToken`.

## Pagination

`GetItemsByParent` and related endpoints take `StartIndex` and `Limit` query params.
