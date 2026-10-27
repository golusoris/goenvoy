# AGENTS — arr/cleanuparr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical repository: <https://github.com/Cleanuparr/Cleanuparr>
- Pinned tag / commit: `v2.10.11` (`9f324ead2a732f3a5ed1ddc36448c661d6b98263`)
- Last verified: 2026-10-16
- Details: [docs/upstream/arr-cleanuparr.md](../../docs/upstream/arr-cleanuparr.md)

## Auth model

API key passed in `X-Api-Key` HTTP header via `arr.BaseClient`.

## Pagination

- `/api/strikes/recent`: queries recent strikes clamped to `[1, 50]` (`count` query parameter).
- `/api/strikes`: page-based pagination with `page` and `pageSize` (normalized to `[1, 500]`, default 50).
- `SummarizeStrikes`: walks pages using `pageSize=500` with hard limit of 4 pages (HISS-02 scalar bound); sets `StrikeSummary.Partial = true` when 4-page bound stops before `TotalPages`.

## Known quirks

- READ-ONLY module: only HTTP GET endpoints are accessed. Any mutating operation is prohibited.
- `GET /api/status` wraps application metadata under `"application"` and media managers under `"mediaManagers"`.
- `RecentStrikeDto.IsDryRun` wire property is `isDryRun`.
- `GET /api/configuration/seeker` returns seeker intervals, proactive search flags, and media manager instances.
- `GET /api/configuration/general` returns general system settings, logging configuration, and authentication rules.
- `GET /api/configuration/{kind}` returns *arr service configuration (`sonarr`, `radarr`, `lidarr`, `readarr`, `whisparr`, `sportarr`, `lazylibrarian`).
- `ArrInstance` is an explicit allow-list without `apiKey` or catch-all fields to prevent secret exposure.
