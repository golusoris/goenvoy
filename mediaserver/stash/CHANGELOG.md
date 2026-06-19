# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates the endpoint URL. Callers must check the error.

## [2.0.0](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v1.2.1...mediaserver/stash/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.2.1](https://github.com/golusoris/goenvoy/compare/mediaserver/stash/v1.2.0...mediaserver/stash/v1.2.1) (2026-04-15)


### Code Refactoring

* move stash/ to mediaserver/stash/ ([54e10dd](https://github.com/golusoris/goenvoy/commit/54e10ddd0a087142fdaf46f01761a9121734366a))
