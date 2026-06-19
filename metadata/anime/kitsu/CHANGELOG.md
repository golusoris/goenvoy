# Changelog

## Unreleased

- Fix data race on concurrent Set*/read of OAuth tokens and client secret.

## [1.3.1](https://github.com/golusoris/goenvoy/compare/metadata/anime/kitsu/v1.3.0...metadata/anime/kitsu/v1.3.1) (2026-06-19)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* guard Set* methods with RWMutex for thread safety ([e67eb43](https://github.com/golusoris/goenvoy/commit/e67eb4397c125441a837b7bf692488e87f2fd1be))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/anime/kitsu/v1.2.0...metadata/anime/kitsu/v1.3.0) (2026-04-15)


### Features

* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))


### Code Refactoring

* restructure metadata into categories + fix error handling ([902861c](https://github.com/golusoris/goenvoy/commit/902861c594e4aad2ae8e249064082fa496be3f5a))
