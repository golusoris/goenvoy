# Changelog

## [2.1.6](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.5...metadata/book/googlebooks/v2.1.6) (2026-10-27)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.3 ([#194](https://github.com/golusoris/goenvoy/issues/194)) ([6054bca](https://github.com/golusoris/goenvoy/commit/6054bcae0791d621d92a038c5d52ad3d090298bb))

## [2.1.5](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.4...metadata/book/googlebooks/v2.1.5) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.1 ([#162](https://github.com/golusoris/goenvoy/issues/162)) ([9182fcd](https://github.com/golusoris/goenvoy/commit/9182fcdbb8b8da4c2a248386237c5f7299287094))

## [2.1.4](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.3...metadata/book/googlebooks/v2.1.4) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.1.3](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.2...metadata/book/googlebooks/v2.1.3) (2026-10-19)

### Bug Fixes

* **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [2.1.2](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.1...metadata/book/googlebooks/v2.1.2) (2026-10-11)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([d7c9c63](https://github.com/golusoris/goenvoy/commit/d7c9c635422bda81c3a2f8703e5573a1c62b217d))
* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [2.1.1](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.1.0...metadata/book/googlebooks/v2.1.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* guard nil search params and encode OpenSubtitles parent filters ([5fe5b7b](https://github.com/golusoris/goenvoy/commit/5fe5b7b87c3c5518376bc1718ec7d80bba2737ee))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/metadata/book/googlebooks/v2.0.0...metadata/book/googlebooks/v2.1.0) (2026-04-15)

### Features

* add 8 new service modules ([5e07309](https://github.com/golusoris/goenvoy/commit/5e0730980c41255ae89a2c6fd00690bcf7430e62))
* **googlebooks:** add bookshelves, annotations, series, and library methods ([99ed5e6](https://github.com/golusoris/goenvoy/commit/99ed5e6563f291a21d886af0572d2ab48fa13408))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))

### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
* resolve golangci-lint CI errors across 6 packages ([1ef79c1](https://github.com/golusoris/goenvoy/commit/1ef79c17f6c98a701f5001405a7686fb35ddc3b9))
