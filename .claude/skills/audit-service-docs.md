---
name: audit-service-docs
description: Refresh every docs/upstream/<service>.md with today's date and HEAD-check the pinned URL.
---

# Skill — `/audit-service-docs`

Walk every `docs/upstream/*.md`, verify the pinned upstream-API URL still resolves, and stamp `Last verified: <today>`.

## When to use

- Quarterly housekeeping.
- Before a flagship release, to confirm no upstream API surface has moved silently.
- When the user says "audit upstream docs" / "refresh service docs".

## Expected arguments

None (optional: `$1` = path glob, default `docs/upstream/*.md`).

## Steps

1. Glob all targets. If none exist, tell the user `docs/upstream/` is empty and stop.
2. For each file:
   - Read the frontmatter. Extract the pinned URL (field: `url:`) and the last-verified date (field: `last_verified:`).
   - `curl -sSLI -o /dev/null -w "%{http_code}" "<url>"` — capture the status code.
   - If the status is `2xx`: update `last_verified:` to today's date (YYYY-MM-DD). Save.
   - If the status is `3xx` with a `Location:` header: follow it once; if the final URL differs from the pinned URL, flag for user review — DO NOT auto-update the `url:` field (could be a dark pattern / sales redirect).
   - If the status is `4xx` or `5xx`: leave `last_verified:` untouched and append a `# BROKEN (<code>): <today>` line under the frontmatter.
3. Collect a summary table: service | status | old date → new date | notes.
4. Report the table to the user. If any row is `BROKEN` or `REDIRECT`, highlight those and recommend opening a tracking issue.
5. If any `last_verified:` was updated, stage `docs/upstream/*.md` and commit with message `docs(upstream): refresh verified dates (N services)`.

## Don't

- Don't silently rewrite a `url:` that 301-redirects — redirect targets are often marketing URLs, login walls, or sales pages that return 200 but don't contain API docs.
- Don't hit the upstream API itself (`/api/v3/...`); only the docs URL. API endpoints commonly return 401/403 for unauthenticated probes.
- Don't use curl with `-s` alone (no `-S`); you'll silently swallow DNS/TLS errors and mislabel them as `000`.
- Don't parallelize aggressively. A modest `--rate-limit` / serial loop is kinder to the docs hosts and less likely to get you rate-limited.
