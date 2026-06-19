# AGENTS — metadata/tracking/trakt

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://trakt.docs.apiary.io/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-tracking-trakt.md](../../../docs/upstream/metadata-tracking-trakt.md)

## Auth model

`Trakt-Api-Key` + `Trakt-Api-Version` headers are always sent; user endpoints add `Authorization: Bearer <token>` obtained via OAuth2 (authorization-code or device-code flow).

## Pagination

Paginated endpoints echo `X-Pagination-Page`, `X-Pagination-Limit`, `X-Pagination-Page-Count`, and `X-Pagination-Item-Count` headers.

## Rate limits

Device-code poll loop honours `429 Too Many Requests` by sleeping for the advertised interval.
