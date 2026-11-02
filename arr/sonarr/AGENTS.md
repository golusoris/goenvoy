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

Manual import: `GetManualImportCandidates` lists candidates with series mapping and rejections; `ManualImportCandidate.ImportFile` maps candidate to `ManualImportFile`; `SendManualImport` queues `ManualImport` command, which imports. `ProcessManualImport` (POST /manualimport) only re-evaluates candidates and imports nothing.

Backup upload: `UploadBackup` uploads and restores the archive via POST `/api/v3/system/backup/restore/upload`, restarting Sonarr; no upload-only endpoint exists.

Root folders: `UpdateRootFolder` is deprecated; Sonarr serves no PUT `/api/v3/rootfolder/{id}` (fails with 404 or 405), delete and re-add instead.
