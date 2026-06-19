# AGENTS — metadata/video/opensubtitles

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://opensubtitles.stoplight.io/docs/opensubtitles-api/b1eb44d4c8502-open-subtitles-api>
- Last verified: 2026-06-14
- Details: [docs/upstream/metadata-video-opensubtitles.md](../../../docs/upstream/metadata-video-opensubtitles.md)

## Auth model

API key sent as `Api-Key` header on every request; user endpoints additionally require a Bearer token obtained via `Login`.

## Pagination

Search endpoints return `page`, `total_pages`, `per_page`, and `total_count`; use the `page` query param to advance.
