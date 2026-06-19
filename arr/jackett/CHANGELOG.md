# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates baseURL. Callers must check the error.

## [2.0.0](https://github.com/golusoris/goenvoy/compare/arr/jackett/v1.3.0...arr/jackett/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/jackett/v1.2.0...arr/jackett/v1.3.0) (2026-04-15)


### Features

* add 8 new service modules (Jackett, NZBHydra2, Spotify, Deezer, IGDB, RAWG, ListenBrainz, Steam) ([2f4f16d](https://github.com/golusoris/goenvoy/commit/2f4f16d34d1b3ffa9ca71d89f9589c2442785a9a))


### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
