---
service: qBittorrent
url: https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)
last_verified: 2026-11-10
---

# Upstream API — qBittorrent

- **Canonical docs:** <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)>
- **Last verified:** 2026-11-10

## What this API does

qBittorrent WebUI API. Cookie-session auth via /api/v2/auth/login.

## Endpoints

- `POST /api/v2/torrents/stop` — stops torrents by hashes (`hashes` param, pipe-delimited or `all`; qBittorrent 5.0+).
- `POST /api/v2/torrents/start` — starts torrents by hashes (`hashes` param, pipe-delimited or `all`; qBittorrent 5.0+).

## Related

- Service module: see `downloadclient/qbit` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
