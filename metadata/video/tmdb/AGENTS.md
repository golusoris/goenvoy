# AGENTS — metadata/video/tmdb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://developer.themoviedb.org/reference/intro/getting-started>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-video-tmdb.md](../../docs/upstream/metadata-video-tmdb.md)

## Auth model

v4 API Read Access Token passed as `Authorization: Bearer <token>` header.

## Pagination

Search and discover endpoints take a `page` query param.
