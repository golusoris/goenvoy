# Changelog

## [1.0.6](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.5...metadata/tracking/letterboxd/v1.0.6) (2026-10-27)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.3 ([#194](https://github.com/golusoris/goenvoy/issues/194)) ([6054bca](https://github.com/golusoris/goenvoy/commit/6054bcae0791d621d92a038c5d52ad3d090298bb))

## [1.0.5](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.4...metadata/tracking/letterboxd/v1.0.5) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.1 ([#162](https://github.com/golusoris/goenvoy/issues/162)) ([9182fcd](https://github.com/golusoris/goenvoy/commit/9182fcdbb8b8da4c2a248386237c5f7299287094))

## [1.0.4](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.3...metadata/tracking/letterboxd/v1.0.4) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.0.3](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.2...metadata/tracking/letterboxd/v1.0.3) (2026-10-19)

### Bug Fixes

- **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

### Code Refactoring

- **letterboxd:** keep tokens and client secret in metadata.OAuthState ([#135](https://github.com/golusoris/goenvoy/issues/135)) ([3cba8a8](https://github.com/golusoris/goenvoy/commit/3cba8a80938cc73ff4ea3c3c8d858a5b9a0bcea6))

## [1.0.2](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.1...metadata/tracking/letterboxd/v1.0.2) (2026-10-11)

### Bug Fixes

- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([d7c9c63](https://github.com/golusoris/goenvoy/commit/d7c9c635422bda81c3a2f8703e5573a1c62b217d))
- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [1.0.1](https://github.com/golusoris/goenvoy/compare/metadata/tracking/letterboxd/v1.0.0...metadata/tracking/letterboxd/v1.0.1) (2026-06-19)

### Bug Fixes

- guard Set* methods with RWMutex for thread safety ([e67eb43](https://github.com/golusoris/goenvoy/commit/e67eb4397c125441a837b7bf692488e87f2fd1be))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/video/letterboxd/v1.2.0...metadata/video/letterboxd/v1.3.0) (2026-04-15)

### Features

- **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- Moved from `metadata/video/letterboxd` to `metadata/tracking/letterboxd`. Update imports accordingly.
- Fix data race on concurrent Set*/read of OAuth tokens and client secret.
