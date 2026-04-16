# AGENTS — metadata/anime/kitsu

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://kitsu.docs.apiary.io/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-anime-kitsu.md](../../docs/upstream/metadata-anime-kitsu.md)

## Auth model

Public endpoints are open; user endpoints use `Authorization: Bearer <token>` obtained via OAuth2 password-grant against `/api/oauth/token`.
