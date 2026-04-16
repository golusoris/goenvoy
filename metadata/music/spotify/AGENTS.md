# AGENTS — metadata/music/spotify

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://developer.spotify.com/documentation/web-api>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-music-spotify.md](../../docs/upstream/metadata-music-spotify.md)

## Auth model

OAuth2 access token passed as `Authorization: Bearer <token>` — the caller supplies a token obtained via client-credentials or authorization-code flow.

## Pagination

Search and list endpoints take `limit` and `offset` query params.
