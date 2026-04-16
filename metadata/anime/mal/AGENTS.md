# AGENTS — metadata/anime/mal

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://myanimelist.net/apiconfig/references/api/v2>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-anime-mal.md](../../docs/upstream/metadata-anime-mal.md)

## Auth model

Guest reads use `X-Mal-Client-Id`; user endpoints use `Authorization: Bearer <token>` obtained via OAuth2 with PKCE.

## Known quirks

- When a Bearer token is set it takes priority over `X-Mal-Client-Id`; the header is not sent alongside the token.
