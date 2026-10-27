---
service: Cleanuparr
url: https://github.com/Cleanuparr/Cleanuparr
last_verified: 2026-10-16
---

# Upstream API — Cleanuparr

- **Canonical source:** <https://github.com/Cleanuparr/Cleanuparr> (tag `v2.10.11`, commit `9f324ead2a732f3a5ed1ddc36448c661d6b98263`)
- **Last verified:** 2026-10-16

## What this API does

Cleanuparr — automated queue cleaner, failed-download striker, and stalled-download management tool for the *arr stack.

## Endpoints

- `GET /api/status` — server runtime status and media manager instance counts.
- `GET /api/strikes/recent` — recent strikes clamped to `[1, 50]`.
- `GET /api/strikes` — paginated download item strikes.
- `GET /api/configuration/seeker` — seeker search intervals, proactive search, and instance configurations.
- `GET /api/configuration/general` — general application settings, logging, and auth rules.
- `GET /api/configuration/{kind}` — *arr service configuration (`sonarr`, `radarr`, `lidarr`, `readarr`, `whisparr`, `sportarr`, `lazylibrarian`).

## Related

- Service module: see `arr/cleanuparr` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
