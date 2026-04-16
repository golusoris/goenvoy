# AGENTS — arr/radarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://radarr.video/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-radarr.md](../../docs/upstream/arr-radarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.
