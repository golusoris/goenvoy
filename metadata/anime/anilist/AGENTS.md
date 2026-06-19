# AGENTS — metadata/anime/anilist

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://docs.anilist.co/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-anime-anilist.md](../../../docs/upstream/metadata-anime-anilist.md)

## Auth model

Anonymous GraphQL queries work unauthenticated; user mutations require an OAuth2 access token passed as `Authorization: Bearer <token>`.

## Known quirks

- AniList uses British spelling (`favourites`) — the schema fragments carry `//nolint:misspell` justifications.
