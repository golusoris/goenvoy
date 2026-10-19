# AGENTS — arr/lidarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://lidarr.audio/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-lidarr.md](../../docs/upstream/arr-lidarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## Pagination

Paged endpoints (queue, history, wanted, blocklist, log) take `page` and `pageSize` query params.

## API notes

Queue management: `GetQueueWithOptions` and `DeleteQueueItemWithOptions` support advanced filtering, nested inclusion, and extended deletion flags.

Manual import: `GetManualImportCandidates` lists candidates with artist and album mapping and rejections; `RescoreManualImport` re-evaluates candidates against a chosen album or release mapping and imports nothing; `ManualImportCandidate.ImportFile` maps candidate to `ManualImportFile`; `SendManualImport` queues `ManualImport` command, which imports. `ProcessManualImport` (POST /manualimport) only re-evaluates candidates and imports nothing.
