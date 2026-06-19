# AGENTS — metadata/tracking/listenbrainz

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://listenbrainz.readthedocs.io/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-tracking-listenbrainz.md](../../../docs/upstream/metadata-tracking-listenbrainz.md)

## Auth model

Read endpoints are open; write endpoints take a user token sent as `Authorization: Token <token>` (use `NewWithToken`).
