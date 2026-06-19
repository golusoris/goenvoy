---
service: LaunchBox
url: https://gamesdb.launchbox-app.com/
last_verified: 2026-06-14
---

# Upstream API — LaunchBox

- **Canonical docs:** <https://gamesdb.launchbox-app.com/>
- **Last verified:** 2026-06-14

## What this API does

LaunchBox Games Database — retro + modern games metadata. The client uses the
public bulk archive at <https://gamesdb.launchbox-app.com/Metadata.zip>; the old
API documentation URL currently returns 404.

## Related

- Service module: see `metadata/game/launchbox` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
