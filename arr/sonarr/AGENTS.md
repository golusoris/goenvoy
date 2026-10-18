# AGENTS — arr/sonarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sonarr.tv/docs/api/>
- Last verified: 2026-04-15
- Details: [docs/upstream/arr-sonarr.md](../../docs/upstream/arr-sonarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## API notes

Queue management: `GetQueueWithOptions` and `DeleteQueueItemWithOptions` support advanced filtering, nested inclusion, and extended deletion flags.

Manual import: `GetManualImportCandidates` lists candidates with their series mapping and rejections; `ManualImportCandidate.ImportFile` maps one to a `ManualImportFile`; `SendManualImport` queues the `ManualImport` command, which imports. `ProcessManualImport` (POST /manualimport) only re-evaluates candidates and imports nothing.
