---
service: Spotify
url: https://developer.spotify.com/documentation/web-api
last_verified: 2026-11-10
---

# Upstream API — Spotify

- **Canonical docs:** <https://developer.spotify.com/documentation/web-api>
- **Last verified:** 2026-11-10

## What this API does

Spotify Web API — tracks, albums, artists, playlists. OAuth2 required.

## Development Mode availability

- **February 2026 changelog:** <https://developer.spotify.com/documentation/web-api/references/changes/february-2026> (checked 2026-11-17).
- Removed for Development Mode Client IDs created on or after February 11, 2026; require extended quota access: `GET /artists/{id}/top-tracks`, `GET /browse/new-releases`, `GET /browse/categories`.
- Existing Development Mode integrations: Spotify postponed this endpoint removal on March 9, 2026. Spotify has not published an updated timeline.
- Not listed under "Endpoints still available" for Development Mode apps; require extended quota access: `GET /artists/{id}/related-artists`, `GET /audio-features/{id}`, `GET /recommendations`.

## Related

- Service module: see `metadata/music/spotify` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
