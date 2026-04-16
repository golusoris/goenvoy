# AGENTS — mediaserver/navidrome

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://www.navidrome.org/docs/developers/subsonic-api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-navidrome.md](../../docs/upstream/mediaserver-navidrome.md)

## Auth model

Subsonic-style token auth: `u`, `t` = md5(password+salt), `s` = salt, `v`, `c`, `f=json` query params on every request.

## Known quirks

- md5 is required by the Subsonic protocol; the `crypto/md5` call carries a `//nolint:gosec` justification.
