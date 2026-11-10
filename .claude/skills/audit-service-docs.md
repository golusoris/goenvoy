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
   - `curl -q -sS -I --proto '=http,https' --max-time 20 -o /dev/null -w '%{http_code}|%header{cf-mitigated}|%{redirect_url}\n' "<url>"` — split the output on `|` with a parser that keeps empty fields (for example `IFS='|' read -r status mitigated target`), never on whitespace. Field 1 is status, field 2 is `cf-mitigated` value (empty if absent), and field 3 is redirect target (empty unless 3xx). Needs curl >= 7.84.0 for `%header{}`. `--proto '=http,https'` makes curl refuse any other scheme with exit `1` before it connects, so a redirect target such as `gopher://` or `file://` is never contacted. Why: the target comes from the upstream server (SSRF, CWE-918). `-q` must stay the first argument: it makes curl skip `.curlrc`, so a `location` setting there cannot make curl follow the redirect (the `3xx` rule needs the unfollowed status).
   - Keep curl's exit status for every probe in this step (the first probe and the `405` GET retry). If it is not `0`, the transfer failed (for example `28` timeout, `7` connection refused) and the fields cannot be trusted: leave `last_verified:` untouched, add no marker, list the row as `UNVERIFIABLE` with `curl exit <n>` in the notes, and apply no other rule. Checked before the challenge rule and the status rules.
   - Field 2 = `challenge`, whatever the status (checked after the exit-status rule and before any status rule): Cloudflare bot wall, not a broken link. Leave `last_verified:` untouched, add no marker, list the row as `UNVERIFIABLE` in the summary.
   - If the status is `2xx`: update `last_verified:` to today's date (YYYY-MM-DD) and remove an existing `broken:` key. Save.
   - If the status is `3xx`: the redirect target is field 3 (no `-L`). If the target in field 3 has a path segment named `login`, `signin` or `sign-in`, it is a login wall: leave `last_verified:` untouched, add no marker, list the row as `UNVERIFIABLE` with the target in the notes. Otherwise probe that target once with the same command; flag the row `REDIRECT` with the target for user review; the row stays `REDIRECT` whatever the target probe returns; if the target probe exits nonzero, add `target probe: curl exit <n>` to the notes — DO NOT auto-update the `url:` field (could be a dark pattern / sales redirect).
   - `HEAD` returns `405`: retry once with the same command minus `-I` (a GET); use the GET fields.
   - other `4xx` / `5xx`: leave `last_verified:` untouched and set frontmatter key `broken: "<code> <today>"` after `last_verified:`, replacing an existing `broken:` key so a file holds at most one. Why: a body `# BROKEN` line is a second H1 and fails markdownlint MD025.
   - Stamp the module `AGENTS.md` `- Last verified:` line in lockstep when it holds only a date; leave lines that carry spec-pin evidence.
3. Collect a summary table: service | status | old date → new date | notes.
4. Report the table to the user. If any row is `BROKEN`, `REDIRECT` or `UNVERIFIABLE`, highlight those and recommend opening a tracking issue for `BROKEN` or `REDIRECT`; `UNVERIFIABLE` rows need a manual browser check, not an issue.
5. If any audit metadata changed (a `last_verified:` stamp, a `broken:` key added, replaced or removed, or a module `AGENTS.md` date stamped), stage the changed `docs/upstream/*.md` and module `AGENTS.md` files and commit with message `docs(upstream): refresh verified dates (N services)`.

## Don't

- Don't silently rewrite a `url:` that 301-redirects — redirect targets are often marketing URLs, login walls, or sales pages that return 200 but don't contain API docs.
- Don't hit the upstream API itself (`/api/v3/...`); only the docs URL. API endpoints commonly return 401/403 for unauthenticated probes.
- Don't use curl with `-s` alone (no `-S`); you'll silently swallow DNS/TLS errors and mislabel them as `000`.
- Don't parallelize aggressively. A modest `--rate-limit` / serial loop is kinder to the docs hosts and less likely to get you rate-limited.
