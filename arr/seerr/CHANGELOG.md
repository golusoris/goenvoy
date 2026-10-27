# Changelog

## [1.6.2](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.6.1...arr/seerr/v1.6.2) (2026-10-27)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.6.2 ([#188](https://github.com/golusoris/goenvoy/issues/188)) ([8e38d2d](https://github.com/golusoris/goenvoy/commit/8e38d2d9ab3b324abdcb21630aed26e8bbcc8ce4))

## [1.6.1](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.6.0...arr/seerr/v1.6.1) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.6.0 ([#158](https://github.com/golusoris/goenvoy/issues/158)) ([241bfd7](https://github.com/golusoris/goenvoy/commit/241bfd7f6f13ec41ac18b28c73776ae40d368a7a))

## [1.6.0](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.5.3...arr/seerr/v1.6.0) (2026-10-23)


### Features

* add watch state, Jellyfin collection writes and Seerr Jellyfin users ([#173](https://github.com/golusoris/goenvoy/issues/173)) ([f653416](https://github.com/golusoris/goenvoy/commit/f65341610b27a2780933aed3bf2dee73937a3c7e))

## [1.5.3](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.5.2...arr/seerr/v1.5.3) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.5.2](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.5.1...arr/seerr/v1.5.2) (2026-10-19)


### Code Refactoring

* **arr:** build the service clients through arr.NewServiceClient ([#137](https://github.com/golusoris/goenvoy/issues/137)) ([f3b7f95](https://github.com/golusoris/goenvoy/commit/f3b7f952b56fc34584869e2ef6ab12e25d843a9a))

## [1.5.1](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.5.0...arr/seerr/v1.5.1) (2026-10-18)

### Code Refactoring

* **arr:** fold the request tail of the remaining arr methods onto the arr/v2 helpers ([#126](https://github.com/golusoris/goenvoy/issues/126)) ([21144ce](https://github.com/golusoris/goenvoy/commit/21144ce9e343c81b8ce667cdc6378fcace46ba91))

## [1.5.0](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.4.0...arr/seerr/v1.5.0) (2026-10-17)

### Features

* **arr/seerr:** tell server 0 from a request without a server ([#104](https://github.com/golusoris/goenvoy/issues/104)) ([3ea04bb](https://github.com/golusoris/goenvoy/commit/3ea04bb89e4abf3175f90d349bcc1555a6f8c276)), closes [#103](https://github.com/golusoris/goenvoy/issues/103)

## [1.4.0](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.3.1...arr/seerr/v1.4.0) (2026-10-17)

### Features

* **arr/seerr:** decode request processing fields and service settings ([#98](https://github.com/golusoris/goenvoy/issues/98)) ([6bf8f0f](https://github.com/golusoris/goenvoy/commit/6bf8f0fb28e04db5ae56f00407d4a39b1beb9b80))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.3.0...arr/seerr/v1.3.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.1.0 ([#64](https://github.com/golusoris/goenvoy/issues/64)) ([273caf4](https://github.com/golusoris/goenvoy/commit/273caf4d61faa26b1d2b9f79e7ca4539025a5f35))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/seerr/v1.2.0...arr/seerr/v1.3.0) (2026-04-15)

### Features

* **arr/seerr:** add Seerr API client ([a3ec03f](https://github.com/golusoris/goenvoy/commit/a3ec03fca77e51a8e316d5f748709f9a70d4da53))

### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
