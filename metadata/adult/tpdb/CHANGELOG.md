# Changelog

## [1.4.0](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.6...metadata/adult/tpdb/v1.4.0) (2026-10-27)


### Features

* **metadata/adult/tpdb:** add performer sites, favourites, changes feed ([#210](https://github.com/golusoris/goenvoy/issues/210)) ([2ad3db2](https://github.com/golusoris/goenvoy/commit/2ad3db2ffe1e858c4344d826176a839453eeb1e3))
* **metadata/adult/tpdb:** add uuid to tags and directors ([#212](https://github.com/golusoris/goenvoy/issues/212)) ([dcaf10b](https://github.com/golusoris/goenvoy/commit/dcaf10bdcbee4f465dcb58e9cbe9fe2a75762b75))


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.3 ([#194](https://github.com/golusoris/goenvoy/issues/194)) ([6054bca](https://github.com/golusoris/goenvoy/commit/6054bcae0791d621d92a038c5d52ad3d090298bb))
* **metadata/adult/tpdb:** decode the uuid id and numeric _id keys ([#211](https://github.com/golusoris/goenvoy/issues/211)) ([4128c4b](https://github.com/golusoris/goenvoy/commit/4128c4b6e0db3d48ab3e6c91791e704976b8f358))

## [1.3.6](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.5...metadata/adult/tpdb/v1.3.6) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.1 ([#162](https://github.com/golusoris/goenvoy/issues/162)) ([9182fcd](https://github.com/golusoris/goenvoy/commit/9182fcdbb8b8da4c2a248386237c5f7299287094))

## [1.3.5](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.4...metadata/adult/tpdb/v1.3.5) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.3.4](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.3...metadata/adult/tpdb/v1.3.4) (2026-10-19)

### Bug Fixes

* **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [1.3.3](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.2...metadata/adult/tpdb/v1.3.3) (2026-10-18)

### Code Refactoring

* **metadata/adult/tpdb:** delegate LookupByHash to FindSceneByHash ([#120](https://github.com/golusoris/goenvoy/issues/120)) ([d4c9758](https://github.com/golusoris/goenvoy/commit/d4c97586df4675a0606e4c2fa9c65a6dcd335607))

## [1.3.2](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.1...metadata/adult/tpdb/v1.3.2) (2026-10-11)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([d7c9c63](https://github.com/golusoris/goenvoy/commit/d7c9c635422bda81c3a2f8703e5573a1c62b217d))
* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.3.0...metadata/adult/tpdb/v1.3.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.0 ([#65](https://github.com/golusoris/goenvoy/issues/65)) ([127fc2d](https://github.com/golusoris/goenvoy/commit/127fc2df0eef17326e41cd0ed1a46011ff3811fe))
* guard nil search params and encode OpenSubtitles parent filters ([5fe5b7b](https://github.com/golusoris/goenvoy/commit/5fe5b7b87c3c5518376bc1718ec7d80bba2737ee))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/adult/tpdb/v1.2.0...metadata/adult/tpdb/v1.3.0) (2026-04-15)

### Features

* add TPDB, MusicBrainz, and Simkl metadata clients ([5098d10](https://github.com/golusoris/goenvoy/commit/5098d1094a6e8ac5cb7dec5fa606a20df1cf3271))
* **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))
