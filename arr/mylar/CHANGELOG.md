# Changelog

## Unreleased

- Use a private HTTP transport pool for the default client, so `CloseIdleConnections` on `http.DefaultTransport` (for example from `httptest.Server.Close`) no longer breaks in-flight requests.
- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.

## [3.0.1](https://github.com/golusoris/goenvoy/compare/arr/mylar/v3.0.0...arr/mylar/v3.0.1) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [3.0.0](https://github.com/golusoris/goenvoy/compare/arr/mylar/v2.1.0...arr/mylar/v3.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.
- constructor sweep — 23 modules return (*Client, error)

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/arr/mylar/v2.0.0...arr/mylar/v2.1.0) (2026-04-15)

### Features

- add 8 new service modules ([5e07309](https://github.com/golusoris/goenvoy/commit/5e0730980c41255ae89a2c6fd00690bcf7430e62))

### Bug Fixes

- BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
