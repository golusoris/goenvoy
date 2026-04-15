# AGENTS — downloadclient

> Per-category conventions for **downloadclient/** modules. Read repo-root [AGENTS.md](../AGENTS.md) + [.workingdir/PRINCIPLES.md](../.workingdir/PRINCIPLES.md) first.

## What this category is

Clients for torrent / Usenet download managers — Deluge, Transmission, qBittorrent, rTorrent, NZBGet, SABnzbd. Used by automation tools (the *arr stack) and custom orchestrators.

## Shared types (in `downloadclient/`)

The category root ([downloadclient/types.go](types.go)) defines common types — `Torrent`, `Tracker`, `FileEntry`, `DownloadClientInfo`. Wire formats differ wildly per service, so these types are the **lowest common denominator** representation returned by each client's listing method.

Service sub-modules **may** import `downloadclient` for the shared types but not each other.

## Shared conventions in this category

- **Interface-ish**: each service exposes roughly the same operations — `ListTorrents`, `AddTorrent(magnet|torrent|url)`, `GetInfo`, `Pause`, `Resume`, `Remove`. There is **no** `interface` declared in the category root — Go interface-satisfaction is implicit; callers can compose one per their needs.
- **Wire variety**:
  - Deluge — JSON-RPC over HTTP POST (`/json` endpoint) + cookie-session auth.
  - NZBGet — JSON-RPC over HTTP POST + basic auth.
  - qBittorrent — custom HTTP + cookie-session auth (`/api/v2/auth/login`).
  - rTorrent — XML-RPC.
  - Transmission — JSON-RPC + `X-Transmission-Session-Id` challenge header.
  - SABnzbd — REST-like HTTP GET with `?apikey=` query param.
- **Pagination**: none — all list endpoints return the full torrent/NZB set.
- **Error body**: per-service. Document in each service's `AGENTS.md`. Always surfaced as `*APIError`.

## Modules in this category

| Module | Purpose | Upstream pin |
|---|---|---|
| [downloadclient/deluge](deluge/) | Deluge (JSON-RPC) | [docs/upstream/downloadclient-deluge.md](../docs/upstream/downloadclient-deluge.md) |
| [downloadclient/transmission](transmission/) | Transmission (JSON-RPC) | [docs/upstream/downloadclient-transmission.md](../docs/upstream/downloadclient-transmission.md) |
| [downloadclient/qbit](qbit/) | qBittorrent WebUI API | [docs/upstream/downloadclient-qbit.md](../docs/upstream/downloadclient-qbit.md) |
| [downloadclient/rtorrent](rtorrent/) | rTorrent (XML-RPC) | [docs/upstream/downloadclient-rtorrent.md](../docs/upstream/downloadclient-rtorrent.md) |
| [downloadclient/nzbget](nzbget/) | NZBGet (JSON-RPC) | [docs/upstream/downloadclient-nzbget.md](../docs/upstream/downloadclient-nzbget.md) |
| [downloadclient/sabnzbd](sabnzbd/) | SABnzbd (HTTP GET API) | [docs/upstream/downloadclient-sabnzbd.md](../docs/upstream/downloadclient-sabnzbd.md) |

## When adding a new service here

Use `/add-service-client downloadclient <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`. Map upstream shapes to the shared types in `downloadclient/types.go` where meaningful.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../_meta/service-agents-template.md)).

For services with cookie-session auth, document the login endpoint + cookie jar behaviour prominently.
