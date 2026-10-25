# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Add WithUserAgent option.
- Add `ProviderIDs` (JSON key `ProviderIds`) to `BaseItemDto`.

## [3.2.0](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v3.1.0...mediaserver/jellyfin/v3.2.0) (2026-10-25)


### Features

* **mediaserver:** add ProviderIDs to Jellyfin and Emby BaseItemDto ([#182](https://github.com/golusoris/goenvoy/issues/182)) ([4e7f84c](https://github.com/golusoris/goenvoy/commit/4e7f84ceba8a209fd57b0e94db2c4e50fd22ebe8))

## [3.1.0](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v3.0.2...mediaserver/jellyfin/v3.1.0) (2026-10-23)


### Features

* add watch state, Jellyfin collection writes and Seerr Jellyfin users ([#173](https://github.com/golusoris/goenvoy/issues/173)) ([f653416](https://github.com/golusoris/goenvoy/commit/f65341610b27a2780933aed3bf2dee73937a3c7e))

## [3.0.2](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v3.0.1...mediaserver/jellyfin/v3.0.2) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [3.0.1](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v3.0.0...mediaserver/jellyfin/v3.0.1) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [3.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v2.1.0...mediaserver/jellyfin/v3.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- **mediaserver:** add WithUserAgent option to 9 modules ([adf9382](https://github.com/golusoris/goenvoy/commit/adf938267b8e32cdc137d9db8571def6512010bf))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/mediaserver/jellyfin/v2.0.0...mediaserver/jellyfin/v2.1.0) (2026-04-15)

### Features

- **mediaserver:** add Jellyfin and Emby clients ([af809d3](https://github.com/golusoris/goenvoy/commit/af809d33c5f3028f1829c7708a72aab94472930d))
