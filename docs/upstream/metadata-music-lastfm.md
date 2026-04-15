---
service: Last.fm
url: https://www.last.fm/api
last_verified: 2026-04-15
---

# Upstream API — Last.fm

- **Canonical docs:** <https://www.last.fm/api>
- **Last verified:** 2026-04-15

## What this API does

Last.fm — scrobbles + artist/track/album metadata. API key + session (for writes).

## Related

- Service module: see `metadata/music/lastfm` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
