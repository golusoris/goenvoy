# AGENTS — arr/jackett

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/Jackett/Jackett/wiki/API-Endpoints-Reference>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-jackett.md](../../docs/upstream/arr-jackett.md)

## Auth model

API key passed as `apikey` query parameter.
Errors never carry key: request-build, transport + response-read errors copied with `apikey=REDACTED` (redactError); unparseable URL or query dropped whole (redactURL); leftover raw or query-escaped key text, decode errors + `APIError.Body` scrubbed (scrub); error whose Unwrap tree (single + joined, first 64 nodes; longer tree hidden whole) holds key text hidden behind opaque wrapper: no `Unwrap`, `errors.Is` + `Timeout` forwarded (redactedError, carriesKey); `New` rejects base URL containing `?` or `#`.
