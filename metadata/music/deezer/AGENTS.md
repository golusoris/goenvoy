# AGENTS — metadata/music/deezer

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://developers.deezer.com/api>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-music-deezer.md](../../docs/upstream/metadata-music-deezer.md)

## Auth model

Public endpoints need no auth; user-specific endpoints take an OAuth2 access token as an `access_token` query parameter (set via `NewWithToken`).
