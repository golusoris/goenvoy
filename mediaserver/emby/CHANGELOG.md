# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Add WithUserAgent option.

## [3.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/emby/v2.1.0...mediaserver/emby/v3.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/mediaserver/emby/v2.0.0...mediaserver/emby/v2.1.0) (2026-04-15)


### Features

* **mediaserver:** add Jellyfin and Emby clients ([af809d3](https://github.com/golusoris/goenvoy/commit/af809d33c5f3028f1829c7708a72aab94472930d))
