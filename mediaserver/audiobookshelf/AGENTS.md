# AGENTS — mediaserver/audiobookshelf

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api.audiobookshelf.org/>
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-audiobookshelf.md](../../docs/upstream/mediaserver-audiobookshelf.md)

## Auth model

Bearer token passed in the `Authorization` header; obtained via `/login` or set explicitly.
