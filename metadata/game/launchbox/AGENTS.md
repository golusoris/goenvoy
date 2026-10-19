# AGENTS — metadata/game/launchbox

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://gamesdb.launchbox-app.com/>
- Last verified: 2026-06-14
- Details: [docs/upstream/metadata-game-launchbox.md](../../../docs/upstream/metadata-game-launchbox.md)

## Auth model

No authentication — client downloads public `Metadata.zip` archive and parses locally.

## Known quirks

- Data source is bulk ZIP download, not REST API; updates infrequent and full dataset loaded in-process.
- Bulk archive: <https://gamesdb.launchbox-app.com/Metadata.zip>.
