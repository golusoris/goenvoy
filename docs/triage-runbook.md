# Downstream-issue triage runbook

> Standing loop for incoming `bug_report.yml` issues. httptest-only reproduction (ADR-0006).
> Owner rotation: `goenvoy-lead` triages + routes; `clients-eng-*` fix/gate/merge; `metadata-scout` scouts only.

## 0. Intake

A downstream bug arrives via `.github/ISSUE_TEMPLATE/bug_report.yml` with fields:
**Module · Description · To Reproduce · Expected · Actual · Go version · module version · target service version.**

If the `Module` field is missing or ambiguous, ask the reporter before proceeding — do not guess the owning module.

## 1. Route to the owning engineer

Map the `Module` field to its category root, then to the responsible engineer:

| Module prefix | Engineer |
|---|---|
| `arr/*` | `clients-eng-arr` |
| `downloadclient/*` | `clients-eng-arr` |
| `mediaserver/*` | `clients-eng-arr` |
| `anime/*` | `clients-eng-arr` |
| `tools/*` | `clients-eng-arr` |
| `metadata/*` (incl. `metadata/{video,anime,music,tracking,adult,book,game}/*`) | `clients-eng-metadata` |

- Only the **coder4090** engineers (`clients-eng-arr`, `clients-eng-metadata`) write, gate, and merge code.
- `metadata-scout` may pre-collect upstream API shapes as **notes only** — it never writes/gates/merges.
- Breaking-change severity escalates to `goenvoy-lead` (see §5).

Assign by creating/assigning a child issue under the bug to the routed engineer.

## 2. Reproduce — httptest ONLY

- **Never hit a live API.** ADR-0006 forbids it; the `guard-go-edit.sh` PreToolUse hook blocks live-API URLs in `*_test.go`.
- Stand up an `httptest.Server` that returns the reporter's captured `Actual` payload/status, point the client at `srv.URL`, and assert the failure. If it doesn't reproduce, request a fuller wire capture before writing a fix.

## 3. Fix in the owning module

Every fix lands in the module named by the `Module` field and must ship the full set:

- **Error wrap**: `fmt.Errorf("<module>: <op>: %w", err)` (`wrapcheck` enforces at boundaries). Never log secrets (rule 7).
- **≥1 table-driven test case** reproducing the bug, driven by `httptest`.
- **A runnable godoc `Example`** for the touched method.
- **`CHANGELOG.md`** — add a line under the module's `Unreleased` section.
- **`docs/upstream/<category>-<service>.md`** — update only if the upstream API surface moved.
- Update the module's `AGENTS.md` if a convention changed.

Preserve the pure-stdlib invariant (ADR-0001): `net/http`, `encoding/json`, `encoding/xml`, `crypto/*`, `context`, `net/url`, `net/http/httptest` only. `//nolint` needs a same-line justification.

## 4. Gate in-pod — PINNED tool versions

Run the per-module gate, then the release check. Use the versions CI pins (`.github/workflows/ci.yml`) so local == CI:

| Tool | Pinned version |
|---|---|
| Go | `stable` (matches `actions/setup-go` `go-version: stable`) |
| golangci-lint | `v2.12.2` |
| gosec | `v2.27.1` |
| govulncheck | `v1.4.0` |
| apidiff | `v0.0.0-20260410095643-746e56fc9e2f` |

```sh
# per-module gate (build · vet · race test · coverage · lint · gosec · govulncheck · apidiff)
tools/release-check.sh <module-path> vX.Y.Z
```

Merge bar (per module): **0 lint · 0 gosec · 0 govulncheck · race-green**.

## 5. Severity → release path

`apidiff` (run by `release-check.sh` against the last `<module>/v*` tag) decides severity:

- **compatible** → patch bump. `/bump-module <module> patch` → `/release-module <module> vX.Y.Z`.
- **breaking** → **escalate to `goenvoy-lead`**, take the `/vN` major-path, and write a `Migration:` footer in the commit body with before/after Go snippets (rule 9). Then `/bump-module <module> major` → `/release-module`.

First release of a module has no prior tag; `release-check.sh` skips apidiff → treat as compatible (patch/minor per SemVer intent).

## 6. Close the loop

Comment the fix + released version on the downstream issue, then close it. Update `.workingdir/STATE.md` if the fix resolved a tracked audit finding.

---

## First-pass triage log

| Date | Open downstream (`bug_report.yml`) issues | Disposition |
|---|---|---|
| 2026-07-04 | **None.** No open issues carry the `bug` label or originate from `bug_report.yml`; the only open board items are governance/release chores (GOE-1..5, `origin=manual`). | Rotation established; nothing to reproduce this pass. |

When the next downstream bug lands, append a row here and route it per §1.
