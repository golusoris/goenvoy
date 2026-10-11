# Changelog

## Unreleased

- Moved from `metadata/video/letterboxd` to `metadata/tracking/letterboxd`. Update imports accordingly.
- Fix data race on concurrent Set*/read of OAuth tokens and client secret.

## [1.0.1](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.0...metadata/tracking/letterboxd/v1.0.1) (2026-06-19)


### Bug Fixes

* guard Set* methods with RWMutex for thread safety ([e67eb43](https://github.com/golusoris/goenvoy/commit/e67eb4397c125441a837b7bf692488e87f2fd1be))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/video/letterboxd/v1.2.0...metadata/video/letterboxd/v1.3.0) (2026-04-15)


### Features

* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))
