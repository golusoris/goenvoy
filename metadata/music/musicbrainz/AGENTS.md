# AGENTS — metadata/music/musicbrainz

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://musicbrainz.org/doc/MusicBrainz_API>
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-music-musicbrainz.md](../../../docs/upstream/metadata-music-musicbrainz.md)

## Auth model

No authentication for read endpoints; callers must set a descriptive `User-Agent` (see `WithUserAgent`).
