# AGENTS — metadata/tracking/simkl

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://simkl.docs.apiary.io/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-tracking-simkl.md](../../docs/upstream/metadata-tracking-simkl.md)

## Auth model

`Simkl-Api-Key` header is always sent; user endpoints add `Authorization: Bearer <token>` obtained via OAuth2 (PIN or authorization-code flow).
