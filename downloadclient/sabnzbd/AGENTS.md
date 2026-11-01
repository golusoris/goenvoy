# AGENTS — downloadclient/sabnzbd

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sabnzbd.org/wiki/advanced/api>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-sabnzbd.md](../../docs/upstream/downloadclient-sabnzbd.md)

## Auth model

API key passed as `apikey` query parameter.

## Known quirks

- Queue priority mutation: `mode=queue`, `name=priority`, `value=<nzo_id>`, `value2=<p>`; returns `{"position": n}`, -1 = unknown job.
