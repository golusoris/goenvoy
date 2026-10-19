# AGENTS — anime

> Per-category conventions for **anime/** modules. Read repo-root [AGENTS.md](../AGENTS.md) + [docs/principles.md](../docs/principles.md) first.

## What this category is

Anime-specific **orchestration / management** services (not catalog-metadata providers — those live under [metadata/anime/](../metadata/anime/)). Currently holds `shoko`; future additions anime-domain but not pure metadata (e.g. scheduling, group-watch, subtitling) land here.

## Shared types (in `anime/`)

Category root ([anime/types.go](types.go)) defines shared types for anime-management concerns. Minimal set currently; add types here if second service duplicates shape from `shoko`.

Service sub-modules **may** import `anime` but not each other.

## Shared conventions in this category

Too few modules today to crystallise category-wide conventions. Each service documents own conventions in its `AGENTS.md`. Expect divergence from upstream APIs for anime services — historically idiosyncratic (AniDB UDP anyone?).

## Modules in this category

| Module | Purpose | Upstream pin |
|---|---|---|
| [anime/shoko](shoko/) | Shoko Server (anime file management) | [docs/upstream/anime-shoko.md](../docs/upstream/anime-shoko.md) |

## When adding a new service here

Use `/add-service-client anime <service> <docs-url>`.

Must-haves at file creation:
- `doc.go` — one-sentence package comment.
- `types.go` — request/response types + `APIError`.
- `<service>.go` — `New` + `Option`s + `do` helper.
- `<service>_test.go` — table-driven tests, `httptest` only.
- `example_test.go` — runnable godoc example.
- `AGENTS.md` — per-service conventions (copy [_meta/service-agents-template.md](../_meta/service-agents-template.md)).

If new service is **metadata provider** (catalog lookup, ratings, images), put under [metadata/anime/](../metadata/anime/) instead.
