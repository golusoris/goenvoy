# AGENTS — metadata/anime/anidb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://wiki.anidb.net/API>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-anime-anidb.md](../../docs/upstream/metadata-anime-anidb.md)

## Auth model

Registered `client` name and `clientver` are sent as query params together with `protover=1`; no user token is required for the HTTP API.
