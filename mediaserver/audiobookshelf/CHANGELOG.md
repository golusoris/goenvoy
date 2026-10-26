# Changelog

## [2.1.0](https://github.com/golusoris/goenvoy/compare/mediaserver/audiobookshelf/v2.0.2...mediaserver/audiobookshelf/v2.1.0) (2026-10-26)


### Features

* **mediaserver/audiobookshelf:** close v2.37.2 gaps in users, auth, pagination and podcasts ([#203](https://github.com/golusoris/goenvoy/issues/203)) ([e24e2a0](https://github.com/golusoris/goenvoy/commit/e24e2a01d1a0e82fee5ac8908e61bfb3d9c3bd93))

## [2.0.2](https://github.com/golusoris/goenvoy/compare/mediaserver/audiobookshelf/v2.0.1...mediaserver/audiobookshelf/v2.0.2) (2026-10-26)


### Bug Fixes

* give every default HTTP client a private transport pool ([#183](https://github.com/golusoris/goenvoy/issues/183)) ([367d63f](https://github.com/golusoris/goenvoy/commit/367d63f8cecc2de3c4fc98a1e3d3f69c11c5080c))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/mediaserver/audiobookshelf/v2.0.0...mediaserver/audiobookshelf/v2.0.1) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/audiobookshelf/v1.3.0...mediaserver/audiobookshelf/v2.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/mediaserver/audiobookshelf/v1.2.0...mediaserver/audiobookshelf/v1.3.0) (2026-04-15)

### Features

- add 8 new service modules ([1dcd7bc](https://github.com/golusoris/goenvoy/commit/1dcd7bce5cc9cb04f5a14a7add647d1f0603f4e0))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Add WithUserAgent option.
