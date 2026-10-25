# Changelog

## [1.3.5](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.3.4...metadata/anime/mal/v1.3.5) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.1 ([#162](https://github.com/golusoris/goenvoy/issues/162)) ([9182fcd](https://github.com/golusoris/goenvoy/commit/9182fcdbb8b8da4c2a248386237c5f7299287094))

## [1.3.4](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.3.3...metadata/anime/mal/v1.3.4) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.3.3](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.3.2...metadata/anime/mal/v1.3.3) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

### Code Refactoring

- **metadata:** fold OAuth state and POST helpers onto metadata v1.4.0 ([#134](https://github.com/golusoris/goenvoy/issues/134)) ([7ea5a20](https://github.com/golusoris/goenvoy/commit/7ea5a2067a139c008360ce3e4b52e0d5c0030b36))

## [1.3.2](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.3.1...metadata/anime/mal/v1.3.2) (2026-10-11)

### Bug Fixes

- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([d7c9c63](https://github.com/golusoris/goenvoy/commit/d7c9c635422bda81c3a2f8703e5573a1c62b217d))
- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.3.0...metadata/anime/mal/v1.3.1) (2026-06-19)

### Bug Fixes

- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
- guard Set* methods with RWMutex for thread safety ([e67eb43](https://github.com/golusoris/goenvoy/commit/e67eb4397c125441a837b7bf692488e87f2fd1be))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/anime/mal/v1.2.0...metadata/anime/mal/v1.3.0) (2026-04-15)

### Features

- **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))

### Code Refactoring

- restructure metadata into categories + fix error handling ([902861c](https://github.com/golusoris/goenvoy/commit/902861c594e4aad2ae8e249064082fa496be3f5a))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- Fix data race on concurrent Set*/read of OAuth tokens and client secret.
