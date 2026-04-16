# AGENTS — mediaserver/kavita

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://wiki.kavitareader.com/guides/misc/api>
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-kavita.md](../../docs/upstream/mediaserver-kavita.md)

## Auth model

Bearer JWT passed in the `Authorization` header; obtained via `/api/Plugin/authenticate` using an API key.
