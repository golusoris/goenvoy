# AGENTS — metadata/game/launchbox

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://gamesdb.launchbox-app.com/api/documentation>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-game-launchbox.md](../../docs/upstream/metadata-game-launchbox.md)

## Auth model

No authentication — the client downloads the public `Metadata.zip` archive and parses it locally.

## Known quirks

- Data source is a bulk ZIP download, not a REST API; updates are infrequent and the full dataset is loaded in-process.
