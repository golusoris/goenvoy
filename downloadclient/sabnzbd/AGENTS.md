# AGENTS — downloadclient/sabnzbd

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sabnzbd.org/wiki/advanced/api>
- Last verified: 2026-11-10
- Details: [docs/upstream/downloadclient-sabnzbd.md](../../docs/upstream/downloadclient-sabnzbd.md)

## Auth model

API key passed as `apikey` query parameter.
Transport, status, decode, application + `New` errors scrubbed: raw, query-escaped, path-escaped (percent-hex case ignored), quoted + JSON-escaped key forms replaced with `REDACTED`. `New` diagnostics drop userinfo; unparseable URL/query drops `?` suffix, prefix kept + scrubbed. Root-only `*url.Error` copy (untouched root keeps identity), `Op` + `URL` scrubbed; key-bearing or >64-node error trees hidden (over-budget tree uses fixed message). `errors.As` yields sanitized `*url.Error` copy (64-node scan); `errors.Is`, `Timeout`, `Temporary` forwarded best-effort inside walk limits (64 joined errors, 4096 visited nodes), separate from 64-node key scan deciding hiding.

## Known quirks

- Queue priority mutation: `mode=queue`, `name=priority`, `value=<nzo_id>`, `value2=<p>`; returns `{"position": n}`, -1 = unknown job.
