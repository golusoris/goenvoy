# AGENTS — metadata/game/screenscraper

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://www.screenscraper.fr/webapi2.php>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-game-screenscraper.md](../../docs/upstream/metadata-game-screenscraper.md)

## Auth model

Developer credentials (`devid`, `devpassword`) passed as query parameters; optional user credentials (`ssid`, `sspassword`) raise per-user quotas.
