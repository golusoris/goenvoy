# Changelog

## Unreleased

- Use a private HTTP transport pool for the default client, so `CloseIdleConnections` on `http.DefaultTransport` (for example from `httptest.Server.Close`) no longer breaks in-flight requests.
- Return an error when closing a response body fails instead of discarding it.
- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Add WithUserAgent option.

## [2.0.2](https://github.com/golusoris/goenvoy/compare/mediaserver/kavita/v2.0.1...mediaserver/kavita/v2.0.2) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/mediaserver/kavita/v2.0.0...mediaserver/kavita/v2.0.1) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [2.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/kavita/v1.3.0...mediaserver/kavita/v2.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/mediaserver/kavita/v1.2.0...mediaserver/kavita/v1.3.0) (2026-04-15)

### Features

- add 8 new service modules ([5e07309](https://github.com/golusoris/goenvoy/commit/5e0730980c41255ae89a2c6fd00690bcf7430e62))

### Bug Fixes

- BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
- TMDb query/path injection, Kavita token race condition ([8be265a](https://github.com/golusoris/goenvoy/commit/8be265a4753653a2960c4e1bb4534efa44cd82c4))
