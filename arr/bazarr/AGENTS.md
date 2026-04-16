# AGENTS — arr/bazarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://wiki.bazarr.media/Additional-Configuration/API/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-bazarr.md](../../docs/upstream/arr-bazarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.
