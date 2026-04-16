# AGENTS — downloadclient/qbit

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-qbit.md](../../docs/upstream/downloadclient-qbit.md)

## Auth model

`Login(username, password)` posts to `/api/v2/auth/login`; the server returns a `SID` cookie which the HTTP client's cookie jar carries on subsequent requests.

## Known quirks

- Every request sets a `Referer` header matching the base URL — qBittorrent rejects requests without it.
