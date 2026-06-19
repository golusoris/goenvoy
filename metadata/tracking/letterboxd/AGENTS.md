# AGENTS — metadata/tracking/letterboxd

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api-docs.letterboxd.com/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-tracking-letterboxd.md](../../../docs/upstream/metadata-tracking-letterboxd.md)

## Auth model

OAuth2 client-credentials flow; the resulting access token is sent as `Authorization: Bearer <token>`.
