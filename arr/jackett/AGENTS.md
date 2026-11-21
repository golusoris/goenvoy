# AGENTS — arr/jackett

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/Jackett/Jackett/wiki/Jackett-API>
- Last verified: 2026-11-10
- Details: [docs/upstream/arr-jackett.md](../../docs/upstream/arr-jackett.md)

## Auth model

API key passed as `apikey` query parameter.
New rejects base URL with `?` or `#`; key only via `apiKey` argument; errors never carry key: root-only `*url.Error` copy on request-build, transport + response-read errors (untouched root keeps identity; `Op` scrubbed, URL `apikey=REDACTED`; redactError); unparseable URL or query -> query dropped, prefix kept + scrubbed (redactURL); `New` diagnostics print neither base URL nor parse cause; raw, query-escaped, path-escaped (percent-escape hex case ignored), quoted + JSON-escaped key forms in decode errors, `APIError.Status` + `APIError.Body` replaced inline with `REDACTED`, other text kept (scrub); error whose Unwrap tree (single + joined, 64-node structural key scan; over-budget or typed-nil node in tree -> fixed message, no `Error()` call, typed-nil skipped in walk; isNilError) holds key text hidden behind opaque wrapper: no `Unwrap`, best-effort `errors.Is`, `Timeout`, `Temporary` forwarding walk (64 joined errors, 4096 visited nodes); `errors.As` for `*url.Error` breadth-first over first 64 queued nodes (copy carries opaque cause; hideCause), deeper URL errors not found; structural key scan breadth-first over 64 nodes (redactedError, carriesKey); any body, JSON-valid or not, with key form left after inline scrub once escapes decode -> whole body `REDACTED` (scrubBody).
