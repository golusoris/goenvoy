# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.
- Moved from `arr/flaresolverr` to `tools/flaresolverr`. Update imports accordingly.

## [2.0.0](https://github.com/golusoris/goenvoy/compare/tools/flaresolverr/v1.0.0...tools/flaresolverr/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* import paths gain a /v2 suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* bump screenscraper and flaresolverr to /v2 (missed in the major sweep) ([8207981](https://github.com/golusoris/goenvoy/commit/8207981eb47230b6b2bf341198fe89d35fd7de3d))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/flaresolverr/v1.2.0...arr/flaresolverr/v1.3.0) (2026-04-15)


### Features

* add 8 new service modules ([5e07309](https://github.com/golusoris/goenvoy/commit/5e0730980c41255ae89a2c6fd00690bcf7430e62))
