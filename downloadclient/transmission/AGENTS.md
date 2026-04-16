# AGENTS — downloadclient/transmission

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/transmission/transmission/blob/main/docs/rpc-spec.md>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-transmission.md](../../docs/upstream/downloadclient-transmission.md)

## Auth model

Optional HTTP Basic (`username`/`password`); the client also auto-negotiates the `X-Transmission-Session-Id` header after a 409 response.

## Known quirks

- A first request returns HTTP 409 with a fresh `X-Transmission-Session-Id`; the client captures it and retries transparently.
