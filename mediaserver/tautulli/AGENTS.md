# AGENTS — mediaserver/tautulli

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/Tautulli/Tautulli/wiki/Tautulli-API-Reference>
- Last verified: 2026-11-10
- Details: [docs/upstream/mediaserver-tautulli.md](../../docs/upstream/mediaserver-tautulli.md)

## Auth model

API key passed as `apikey` query parameter against `/api/v2`.
Errors never carry key: root-only `*url.Error` copy on request-build, transport + response-read errors (untouched root keeps identity; `Op` scrubbed, URL `apikey=REDACTED`; redactError); unparseable URL or query -> query dropped, prefix kept + scrubbed (redactURL); `New` diagnostics drop userinfo (redactBaseURL); raw, query-escaped, path-escaped (percent-escape hex case ignored), quoted + JSON-escaped key forms in `New` errors, decode errors, `APIError.Status` + `APIError.Body` scrubbed (scrub); error whose Unwrap tree (single + joined, 64-node structural key scan; over-budget tree -> fixed message, no `Error()` call) holds key text hidden behind opaque wrapper: no `Unwrap`, best-effort `errors.Is`, `Timeout`, `Temporary` + `*url.Error` `errors.As` forwarded inside walk limits (64 joined errors, 4096 visited nodes, separate from 64-node key scan; redactedError, carriesKey).
