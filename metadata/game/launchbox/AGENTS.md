# AGENTS — metadata/game/launchbox

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://gamesdb.launchbox-app.com/>
- Last verified: 2026-06-14
- Details: [docs/upstream/metadata-game-launchbox.md](../../../docs/upstream/metadata-game-launchbox.md)

## Auth model

No authentication — the client downloads the public `Metadata.zip` archive and parses it locally.

## Known quirks

- Data source is a bulk ZIP download, not a REST API; updates are infrequent and the full dataset is loaded in-process.
- Bulk archive: <https://gamesdb.launchbox-app.com/Metadata.zip>.
