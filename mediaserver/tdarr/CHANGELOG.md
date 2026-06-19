# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates the base URL. Callers must check the error.
- Add WithUserAgent option.

## [3.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/tdarr/v2.1.0...mediaserver/tdarr/v3.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/mediaserver/tdarr/v2.0.0...mediaserver/tdarr/v2.1.0) (2026-04-15)


### Features

* add 8 new service modules ([5e07309](https://github.com/golusoris/goenvoy/commit/5e0730980c41255ae89a2c6fd00690bcf7430e62))
