# AGENTS — arr/mylar

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/mylar3/mylar3/wiki/API-Documentation>
- Last verified: 2026-11-10
- Details: [docs/upstream/arr-mylar.md](../../docs/upstream/arr-mylar.md)

## Auth model

API key passed as `apikey` query parameter.
Errors never carry key: root-only `*url.Error` copy on request-build, transport + response-read errors (untouched root keeps identity; `Op` scrubbed, URL `apikey=REDACTED`; redactError); unparseable URL or query dropped whole (redactURL); raw, query-escaped, path-escaped, lower-case percent-hex, quoted + JSON-escaped key forms in `New` errors, decode errors, `APIError.Status` + `APIError.Body` scrubbed (scrub); error whose Unwrap tree (single + joined, first 64 nodes; longer tree hidden whole) holds key text hidden behind opaque wrapper: no `Unwrap`, `errors.Is`, `Timeout` + `Temporary` forwarded (redactedError, carriesKey).
