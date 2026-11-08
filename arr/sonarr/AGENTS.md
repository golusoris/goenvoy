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

Backup restore: `UploadBackup` (POST `/api/v3/system/backup/restore/upload`) and `RestoreBackup` overwrite config.xml at once, stage archive's SQLite database (`sonarr.db` / `nzbdrone.db`) for next start, never restart Sonarr -> call `Restart` after (measured 4.0.20, 2026-11-08). Archive without database (PostgreSQL instance backup) restores config.xml only; earlier staged database stays pending (source 4.0.20.3014). Uploads share one temp path on Sonarr: no retry while earlier attempt may still run. No upload-only endpoint exists.

Root folders: `UpdateRootFolder` is deprecated; Sonarr serves no PUT `/api/v3/rootfolder/{id}` (fails with 404 or 405), delete and re-add instead.
