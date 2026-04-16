# AGENTS — arr/sonarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sonarr.tv/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-sonarr.md](../../docs/upstream/arr-sonarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.
