# AGENTS — metadata/tracking/listenbrainz

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://listenbrainz.readthedocs.io/en/latest/>
- Last verified: 2026-11-10
- Details: [docs/upstream/metadata-tracking-listenbrainz.md](../../../docs/upstream/metadata-tracking-listenbrainz.md)

## Auth model

Read endpoints are open; write endpoints take a user token sent as `Authorization: Token <token>` (use `NewWithToken`).
