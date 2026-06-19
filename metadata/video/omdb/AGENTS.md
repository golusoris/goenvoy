# AGENTS — metadata/video/omdb

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://www.omdbapi.com/>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-video-omdb.md](../../../docs/upstream/metadata-video-omdb.md)

## Auth model

API key passed as `apikey` query parameter.

## Known quirks

- Returns HTTP 200 with `{"Response":"False","Error":"..."}` on lookup misses; the client surfaces this as an `APIError`.
