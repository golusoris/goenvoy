# AGENTS — mediaserver/audiobookshelf

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://api.audiobookshelf.org/>
- Pinned server: v2.37.2 (commit `edb1f7efc92ecacdef51aafb4d9d5054ce10604f`)
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-audiobookshelf.md](../../docs/upstream/mediaserver-audiobookshelf.md)

## Auth model

- Bearer token in `Authorization: Bearer <token>` header.
- Auth exception: `New(baseURL, token, opts...)` accepts empty `token` (`""`) so unauthenticated callers can invoke `Login(ctx, username, password)` (`POST /login` with `X-Return-Tokens: true`) or `GetStatus(ctx)` (`GET /status`).
- `Logout(ctx)` posts to `/logout`.

## Root routes vs `/api` routes

- Root routes (no `/api` prefix, mounted in `server/Server.js`): `POST /login`, `POST /logout`, `GET /status`.
- All other endpoints use `/api` prefix (`server/routers/ApiRouter.js`).
- `GetServerInfo` (`GET /api/server`) is deprecated because Audiobookshelf v2.37.2 has no `/api/server` route; use `GetStatus` (`GET /status`).

## Pagination

- `GetLibraryItemsWithOptions` (`GET /api/libraries/:id/items`) supports query parameters `limit`, `page`, `sort`, `desc` (`0`/`1`), `filter`, `minified` (`0`/`1`), `collapseseries` (`0`/`1`), and `include`.
