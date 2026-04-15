---
service: Transmission
url: https://github.com/transmission/transmission/blob/main/docs/rpc-spec.md
last_verified: 2026-04-15
---

# Upstream API — Transmission

- **Canonical docs:** <https://github.com/transmission/transmission/blob/main/docs/rpc-spec.md>
- **Last verified:** 2026-04-15

## What this API does

Transmission RPC — JSON-RPC with CSRF challenge via X-Transmission-Session-Id.

## Related

- Service module: see `downloadclient/transmission` in the monorepo tree.
- Internal notes: the service's `AGENTS.md` lists auth model, pagination, and known quirks.

> Refreshed by `/audit-service-docs`. Do not hand-edit `last_verified` — run the skill.
