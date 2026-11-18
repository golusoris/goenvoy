# AGENTS — downloadclient/sabnzbd

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sabnzbd.org/wiki/advanced/api>
- Last verified: 2026-11-10
- Details: [docs/upstream/downloadclient-sabnzbd.md](../../docs/upstream/downloadclient-sabnzbd.md)

## Auth model

API key passed as `apikey` query parameter.
Transport, status, decode + application errors scrubbed: raw, query-escaped, path-escaped (percent-hex case ignored), quoted + JSON-escaped key forms replaced with `REDACTED`. `New` diagnostics print neither base URL nor parse cause (fixed message); unparseable `*url.Error` URL/query drops `?` suffix, scrubbed before + after `redactURL`. Root-only `*url.Error` copy (untouched root keeps identity), `Op` + `URL` scrubbed; key-bearing, typed-nil-holding or >64-node error trees hidden (typed-nil or over-budget tree -> fixed message, typed-nil skipped in walk). `errors.As` yields sanitized `*url.Error` copy (64-node scan); `errors.Is`, `Timeout`, `Temporary` forwarded best-effort inside walk limits (64 joined errors, 4096 visited nodes), separate from 64-node key scan deciding hiding.

## Known quirks

- Queue priority mutation: `mode=queue`, `name=priority`, `value=<nzo_id>`, `value2=<p>`; returns `{"position": n}`, -1 = unknown job.
