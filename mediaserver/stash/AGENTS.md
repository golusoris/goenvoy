# AGENTS — mediaserver/stash

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://docs.stashapp.cc/in-app-manual/>
- Pinned schema: v0.31.1 (commit 4de2351e7cc990d7ccd7cb6c84c275cd53bf6e55)
- Last verified: 2026-04-15
- Details: [docs/upstream/mediaserver-stash.md](../../docs/upstream/mediaserver-stash.md)

## Auth model

Optional API key -> `Apikey` header on GraphQL endpoint.

## Pagination

GraphQL filter input: `per_page` + page cursor.

## Mutations

Write-back mutations:

- `PerformerCreate(ctx, input)`: `performerCreate(input: PerformerCreateInput!): Performer`
- `PerformerUpdate(ctx, input)`: `performerUpdate(input: PerformerUpdateInput!): Performer`
- `SceneUpdate(ctx, input)`: `sceneUpdate(input: SceneUpdateInput!): Scene` (carries `performer_ids`)

- Empty required field (`PerformerCreateInput.Name`, `PerformerUpdateInput.ID`, `SceneUpdateInput.ID`) -> error before HTTP.
- Nil pointer field -> omitted (absent = unchanged).
- Slice fields `omitzero`: nil -> omitted; empty non-nil -> sent as `[]`, clears field.
- `schema_fields_test.go` pins every input's JSON keys to v0.31.1 schema.
