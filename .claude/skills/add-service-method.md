---
name: add-service-method
description: Add a single typed method + test case to an existing goenvoy API client.
---

# Skill — `/add-service-method`

Add one typed method to an existing `<category>/<service>` client module.

## When to use

The user says "add a <verb>X method to <service>" and the client already exists. For brand-new modules, use `/add-service-client` first.

## Expected arguments

- `$1` — module path relative to repo root, e.g. `arr/sonarr`, `metadata/video/tmdb`.
- `$2` — method name (Go PascalCase), e.g. `GetSeriesByID`, `ListDownloadClients`.
- `$3` — upstream endpoint (method + path), e.g. `GET /api/v3/series/{id}`.
- (optional) `$4` — link to the upstream API doc anchor for this endpoint.

## Steps

1. `cd <module>` and read `<service>.go`, `types.go`, and existing tests to match the module's conventions (receiver name, header style, URL-building helper, error-wrapping style).
2. Add request/response structs to `types.go`. Name them `<Method>Request` (if the endpoint takes a body or non-trivial query) and `<Method>Response`. Use JSON/XML tags matching the upstream exactly.
3. Implement the method in `<service>.go`:
   - Signature: `func (c *Client) <Method>(ctx context.Context, <pathParams>, <queryParams>, <body>) (<Response>, error)`.
   - First parameter is always `ctx context.Context`.
   - URL-escape any user-provided path segments with `url.PathEscape`.
   - Call the module's `do` / helper; do not build `*http.Request` by hand if a helper exists.
   - Return typed errors via `*APIError` for non-2xx responses.
4. Extend the existing table-driven test in `<service>_test.go`:
   - Add a case to the `tests` slice with `name`, `mockStatus`, `mockBody`, `wantErr`, `want`.
   - The test MUST use `httptest.NewServer`. No live-API URLs. (The `guard-go-edit.sh` hook will block live hosts.)
5. If the new method exposes a new public surface that needs illustration, add a `func Example<Method>()` in `example_test.go`.
6. Update `<module>/AGENTS.md` if the endpoint changes a documented convention (pagination, auth, pagination cursor).
7. Add a bullet to the module's `## Unreleased` section in the root `CHANGELOG.md`:
   - `- **feat:** add <Method> (<endpoint>)`.
8. Run the gate: `go build ./... && go vet ./... && go test -race -count=1 ./... && golangci-lint run --config ../../.golangci.yml ./...`.
9. Commit with a Conventional-Commits message: `feat(<module>): add <Method>`.

## Don't

- Don't concatenate user input into URL paths without `url.PathEscape`.
- Don't add query parameters by string concatenation — use `url.Values` and `u.RawQuery = v.Encode()`.
- Don't forget `defer resp.Body.Close()` if you're handling the response manually instead of going through the module's helper.
- Don't make the response type an interface. Return concrete structs; callers can wrap.
