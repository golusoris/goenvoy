# AGENTS — downloadclient/qbit

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-qbit.md](../../docs/upstream/downloadclient-qbit.md)

## Auth model

`Login(username, password)` posts to `/api/v2/auth/login`; server returns `SID` cookie which HTTP client cookie jar carries on subsequent requests.

## Known quirks

- Every request sets `Referer` header matching base URL — qBittorrent rejects requests without it.
