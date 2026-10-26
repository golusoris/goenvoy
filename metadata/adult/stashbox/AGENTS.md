# AGENTS — metadata/adult/stashbox

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/stashapp/stash-box>
- Pinned schema: v0.10.3 (commit 5eb3551e19fd494e93be410d8149adc785cffd6a)
- Last verified: 2026-04-15
- Details: [docs/upstream/metadata-adult-stashbox.md](../../../docs/upstream/metadata-adult-stashbox.md)

## Auth model

API key -> `Apikey` header on GraphQL endpoint.

## Mutations

Write mutations (schema.graphql, `@hasRole(role: READ)`):

- `FavoritePerformer(ctx, id, favorite)`: `favoritePerformer(id: ID!, favorite: Boolean!): Boolean!`
- `FavoriteStudio(ctx, id, favorite)`: `favoriteStudio(id: ID!, favorite: Boolean!): Boolean!`

- Empty id -> `ErrEmptyID` before HTTP.
- `favorite` always sent, `false` included (unfavorite).
- Null or absent result -> `ErrMissingResult` (schema says non-null).
