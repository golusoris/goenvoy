# Changelog

## [1.3.5](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.3.4...arr/bazarr/v1.3.5) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.6.0 ([#158](https://github.com/golusoris/goenvoy/issues/158)) ([241bfd7](https://github.com/golusoris/goenvoy/commit/241bfd7f6f13ec41ac18b28c73776ae40d368a7a))

## [1.3.4](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.3.3...arr/bazarr/v1.3.4) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.3.3](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.3.2...arr/bazarr/v1.3.3) (2026-10-19)


### Code Refactoring

* **arr:** build the service clients through arr.NewServiceClient ([#137](https://github.com/golusoris/goenvoy/issues/137)) ([f3b7f95](https://github.com/golusoris/goenvoy/commit/f3b7f952b56fc34584869e2ef6ab12e25d843a9a))

## [1.3.2](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.3.1...arr/bazarr/v1.3.2) (2026-10-18)

### Code Refactoring

* **arr:** fold the request tail of the remaining arr methods onto the arr/v2 helpers ([#126](https://github.com/golusoris/goenvoy/issues/126)) ([21144ce](https://github.com/golusoris/goenvoy/commit/21144ce9e343c81b8ce667cdc6378fcace46ba91))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.3.0...arr/bazarr/v1.3.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.1.0 ([#64](https://github.com/golusoris/goenvoy/issues/64)) ([273caf4](https://github.com/golusoris/goenvoy/commit/273caf4d61faa26b1d2b9f79e7ca4539025a5f35))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/bazarr/v1.2.0...arr/bazarr/v1.3.0) (2026-04-15)

### Features

* **arr/bazarr:** add Bazarr API client ([5469769](https://github.com/golusoris/goenvoy/commit/5469769c29c8ab4497cb1c3e2445b572f70eec48))

### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
