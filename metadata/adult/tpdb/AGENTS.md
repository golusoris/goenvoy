# AGENTS — metadata/adult/tpdb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://theporndb.net/>
- Last verified: 2026-06-14
- Details: [docs/upstream/metadata-adult-tpdb.md](../../../docs/upstream/metadata-adult-tpdb.md)

## Auth model

Personal API token passed as `Authorization: Bearer <token>`.

## Known quirks

- The old API docs URL returns 404; client behavior is based on the current
  service API surface until upstream publishes stable docs again.
