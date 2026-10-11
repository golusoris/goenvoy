# Changelog

## Unreleased

- Fix data race on concurrent Set*/read of OAuth tokens and client secret.

## [1.3.1](https://github.com/golusoris/goenvoy/compare/metadata/tracking/simkl/v1.3.0...metadata/tracking/simkl/v1.3.1) (2026-06-19)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* guard Set* methods with RWMutex for thread safety ([e67eb43](https://github.com/golusoris/goenvoy/commit/e67eb4397c125441a837b7bf692488e87f2fd1be))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/tracking/simkl/v1.2.0...metadata/tracking/simkl/v1.3.0) (2026-04-15)


### Features

* add TPDB, MusicBrainz, and Simkl metadata clients ([5098d10](https://github.com/golusoris/goenvoy/commit/5098d1094a6e8ac5cb7dec5fa606a20df1cf3271))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))


### Bug Fixes

* resolve golangci-lint CI errors across 6 packages ([1ef79c1](https://github.com/golusoris/goenvoy/commit/1ef79c17f6c98a701f5001405a7686fb35ddc3b9))
