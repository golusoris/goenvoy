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

Backup restore: `UploadBackup` (POST `/api/v3/system/backup/restore/upload`) and `RestoreBackup` share one Sonarr code path (source 4.0.20.3014). `.zip` extracts into one fixed temp folder, never cleared first; Sonarr takes only `Config.xml` (overwrites config.xml at once) + `sonarr.db` / `nzbdrone.db` (staged for next start); such files left by failed restore apply on next restore. Archive without database (PostgreSQL instance backup) stages none; earlier staged database stays pending unless leftover database from failed restore replaces it. `.db` / `.xml` upload name skips extraction: file itself staged as database -> config.xml only inside `.zip`. Never restarts Sonarr -> call `Restart` after (measured 4.0.20, 2026-11-08). Uploads share one temp path: no retry while earlier attempt may still run. No upload-only endpoint exists.

Root folders: `UpdateRootFolder` is deprecated; Sonarr serves no PUT `/api/v3/rootfolder/{id}` (fails with 404 or 405), delete and re-add instead.
