# AGENTS — downloadclient/deluge

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://deluge.readthedocs.io/en/latest/reference/api.html>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-deluge.md](../../docs/upstream/downloadclient-deluge.md)

## Auth model

`Login(password)` sends a JSON-RPC `auth.login` request; the session cookie returned by the Web UI is carried by the HTTP client's cookie jar.
