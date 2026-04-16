# AGENTS — metadata/game/igdb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api-docs.igdb.com/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-game-igdb.md](../../docs/upstream/metadata-game-igdb.md)

## Auth model

Twitch OAuth2 client-credentials flow; requests carry both `Client-Id` and `Authorization: Bearer <access_token>` headers.

## Known quirks

- Query bodies are Apicalypse text, so the client sets `Content-Type: text/plain`, not JSON.
