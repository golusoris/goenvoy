# AGENTS — metadata/video/tvdb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://thetvdb.github.io/v4-api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-video-tvdb.md](../../docs/upstream/metadata-video-tvdb.md)

## Auth model

API key exchanged via `/login` for a JWT; the JWT is then sent as `Authorization: Bearer <token>`.

## Pagination

Listing endpoints take a zero-based `page` query param.
