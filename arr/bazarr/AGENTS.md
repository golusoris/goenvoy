# AGENTS — arr/bazarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://wiki.bazarr.media/>
- Last verified: 2026-06-14
- Details: [docs/upstream/arr-bazarr.md](../../docs/upstream/arr-bazarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Known quirks

- The old API-specific wiki page returns 404; client behavior is based on
  Bazarr's live `/api` endpoints and the public wiki.
