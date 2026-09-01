# Changelog

## [3.0.0](https://github.com/golusoris/goenvoy/compare/metadata/game/screenscraper/v2.0.1...metadata/game/screenscraper/v3.0.0) (2026-09-01)


### ⚠ BREAKING CHANGES

* import paths gain a /v2 suffix.
* **screenscraper:** unify option type from ...any to typed Option

### Features

* add 6 game metadata API clients ([daf1c3c](https://github.com/golusoris/goenvoy/commit/daf1c3c7f4feefb2712ffa62c7092b6b09d4b7de))
* bump screenscraper and flaresolverr to /v2 (missed in the major sweep) ([8207981](https://github.com/golusoris/goenvoy/commit/8207981eb47230b6b2bf341198fe89d35fd7de3d))


### Bug Fixes

* add go.sum files for new game metadata modules ([a3ec785](https://github.com/golusoris/goenvoy/commit/a3ec7851ec7adfd67954ac54c105a59ad020a31a))
* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))
* guard nil search params and encode OpenSubtitles parent filters ([5fe5b7b](https://github.com/golusoris/goenvoy/commit/5fe5b7b87c3c5518376bc1718ec7d80bba2737ee))
* **screenscraper:** unify option type from ...any to typed Option ([7f161dd](https://github.com/golusoris/goenvoy/commit/7f161dda6747950b87ffe9c174d62e8adf1d3c5e))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/metadata/game/screenscraper/v2.0.0...metadata/game/screenscraper/v2.0.1) (2026-08-31)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [2.0.0](https://github.com/golusoris/goenvoy/compare/metadata/game/screenscraper/v1.0.0...metadata/game/screenscraper/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* import paths gain a /v2 suffix.
* **screenscraper:** unify option type from ...any to typed Option

### Features

* bump screenscraper and flaresolverr to /v2 (missed in the major sweep) ([8207981](https://github.com/golusoris/goenvoy/commit/8207981eb47230b6b2bf341198fe89d35fd7de3d))


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* guard nil search params and encode OpenSubtitles parent filters ([5fe5b7b](https://github.com/golusoris/goenvoy/commit/5fe5b7b87c3c5518376bc1718ec7d80bba2737ee))
* **screenscraper:** unify option type from ...any to typed Option ([7f161dd](https://github.com/golusoris/goenvoy/commit/7f161dda6747950b87ffe9c174d62e8adf1d3c5e))

## Changelog

## Unreleased

- Unify option type: `New` now takes `...screenscraper.Option` instead of `...any`. `metadata.With*` callers should switch to the re-exported `screenscraper.With*` equivalents.
