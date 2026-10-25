# Changelog

## [2.0.2](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v2.0.1...mediaserver/stash/v2.0.2) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v2.0.0...mediaserver/stash/v2.0.1) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

### Code Refactoring

- **mediaserver/stash:** fold by-ID finders into one generic helper ([#140](https://github.com/golusoris/goenvoy/issues/140)) ([f13b03d](https://github.com/golusoris/goenvoy/commit/f13b03de8a22bd90b5b3950fef98e8984537e972))

## [2.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v1.2.1...mediaserver/stash/v2.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.
- constructor sweep — 23 modules return (*Client, error)

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.2.1](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v1.2.0...mediaserver/stash/v1.2.1) (2026-04-15)

### Code Refactoring

- move stash/ to mediaserver/stash/ ([54e10dd](https://github.com/golusoris/goenvoy/commit/54e10ddd0a087142fdaf46f01761a9121734366a))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- **BREAKING:** `New` now returns `(*Client, error)` and validates the endpoint URL. Callers must check the error.
