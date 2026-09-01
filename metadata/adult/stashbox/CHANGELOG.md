# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates the endpoint URL. Callers must check the error.

## [3.0.0](https://github.com/golusoris/goenvoy/compare/metadata/adult/stashbox/v2.0.1...metadata/adult/stashbox/v3.0.0) (2026-09-01)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* add StashBox and Stash GraphQL clients ([c01437d](https://github.com/golusoris/goenvoy/commit/c01437d478ea7e8be045b11cd04440d51f750318))
* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/metadata/adult/stashbox/v2.0.0...metadata/adult/stashbox/v2.0.1) (2026-08-31)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [2.0.0](https://github.com/golusoris/goenvoy/compare/metadata/adult/stashbox/v1.3.0...metadata/adult/stashbox/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/adult/stashbox/v1.2.0...metadata/adult/stashbox/v1.3.0) (2026-04-15)


### Features

* add StashBox and Stash GraphQL clients ([c01437d](https://github.com/golusoris/goenvoy/commit/c01437d478ea7e8be045b11cd04440d51f750318))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))
