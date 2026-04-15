# AGENTS — <category>/<service>

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <URL>
- Pinned version / commit: <semver / date>
- Last verified: <YYYY-MM-DD>

## Auth model

<e.g. "X-Api-Key header"; "Bearer JWT, rotated via /refresh"; "OAuth2 device code">

## Pagination

<e.g. "cursor-based — nextCursor in JSON envelope"; "Link header with RFC 5988"; "none — full list returned">

## Rate limits

<e.g. "5 req/s per API key, HTTP 429 on exceed"; "none documented">

## Known quirks

- <concrete odd behaviour, e.g. "API returns empty array with 200 for unauthenticated calls to /tag">
- <...>

## Testing notes

<e.g. "JSON responses are camelCase except /status which is snake_case; fixture files under testdata/ reflect both">
