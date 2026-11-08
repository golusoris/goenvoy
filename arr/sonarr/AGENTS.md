# AGENTS — arr/sonarr

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://sonarr.tv/docs/api/>
- Last verified: 2026-11-06
- Details: [docs/upstream/arr-sonarr.md](../../docs/upstream/arr-sonarr.md)

## Auth model

API key passed as `X-Api-Key` header via `arr.BaseClient`.

## API notes

Queue management: `GetQueueWithOptions` and `DeleteQueueItemWithOptions` support advanced filtering, nested inclusion, and extended deletion flags.

Manual import: `GetManualImportCandidates` lists candidates with series mapping and rejections; `ManualImportCandidate.ImportFile` maps candidate to `ManualImportFile`; `SendManualImport` queues `ManualImport` command, which imports. `ProcessManualImport` (POST /manualimport) only re-evaluates candidates and imports nothing.

Backup restore: `UploadBackup` (POST `/api/v3/system/backup/restore/upload`) and `RestoreBackup` stage the restore; Sonarr writes config.xml at once, applies database on next start and never restarts itself (measured 4.0.20, 2026-11-08), so call `Restart` after. No upload-only endpoint exists.

Root folders: `UpdateRootFolder` is deprecated; Sonarr serves no PUT `/api/v3/rootfolder/{id}` (fails with 404 or 405), delete and re-add instead.
