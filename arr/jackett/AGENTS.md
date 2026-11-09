# AGENTS — arr/jackett

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/Jackett/Jackett/wiki/API-Endpoints-Reference>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-jackett.md](../../docs/upstream/arr-jackett.md)

## Auth model

API key passed as `apikey` query parameter.
Transport and request-build errors replace the `apikey` query value with `REDACTED` (redactURLError in jackett.go); keep key out of every error.
