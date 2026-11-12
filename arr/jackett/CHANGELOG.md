# Changelog

## [2.0.5](https://github.com/golusoris/goenvoy/compare/arr/jackett/v2.0.4...arr/jackett/v2.0.5) (2026-11-12)


### Bug Fixes

* **deps:** move to Go 1.26.11 for the net/http advisory GO-2026-6859 ([#310](https://github.com/golusoris/goenvoy/issues/310)) ([eee19f2](https://github.com/golusoris/goenvoy/commit/eee19f20da04468690d04e67ed870cf85506218e))
* **jackett:** hide transport errors whose wrapped chain carries the key ([#291](https://github.com/golusoris/goenvoy/issues/291)) ([406cf15](https://github.com/golusoris/goenvoy/commit/406cf15b3b07c975f07265cb2f7aff630e3cb406))
* **jackett:** keep the API key out of errors ([#289](https://github.com/golusoris/goenvoy/issues/289)) ([8088237](https://github.com/golusoris/goenvoy/commit/8088237463cde67a35316f93f167c96b721defb4))

## [2.0.4](https://github.com/golusoris/goenvoy/compare/arr/jackett/v2.0.3...arr/jackett/v2.0.4) (2026-11-08)


### Bug Fixes

* handle error returns that were assigned to blank ([#278](https://github.com/golusoris/goenvoy/issues/278)) ([aa1f0dc](https://github.com/golusoris/goenvoy/commit/aa1f0dce6bf56f7fe65dcdcd782e8fa20a6f3e29))

## [2.0.3](https://github.com/golusoris/goenvoy/compare/arr/jackett/v2.0.2...arr/jackett/v2.0.3) (2026-10-26)


### Bug Fixes

* give every default HTTP client a private transport pool ([#183](https://github.com/golusoris/goenvoy/issues/183)) ([367d63f](https://github.com/golusoris/goenvoy/commit/367d63f8cecc2de3c4fc98a1e3d3f69c11c5080c))

## [2.0.2](https://github.com/golusoris/goenvoy/compare/arr/jackett/v2.0.1...arr/jackett/v2.0.2) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/arr/jackett/v2.0.0...arr/jackett/v2.0.1) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [2.0.0](https://github.com/golusoris/goenvoy/compare/arr/jackett/v1.3.0...arr/jackett/v2.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.
- constructor sweep — 23 modules return (*Client, error)

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/jackett/v1.2.0...arr/jackett/v1.3.0) (2026-04-15)

### Features

- add 8 new service modules (Jackett, NZBHydra2, Spotify, Deezer, IGDB, RAWG, ListenBrainz, Steam) ([2f4f16d](https://github.com/golusoris/goenvoy/commit/2f4f16d34d1b3ffa9ca71d89f9589c2442785a9a))

### Bug Fixes

- BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
