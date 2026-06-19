# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Add WithUserAgent option.

## [2.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/komga/v1.3.0...mediaserver/komga/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/mediaserver/komga/v1.2.0...mediaserver/komga/v1.3.0) (2026-04-15)


### Features

* add 8 new service modules ([1dcd7bc](https://github.com/golusoris/goenvoy/commit/1dcd7bce5cc9cb04f5a14a7add647d1f0603f4e0))


### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
