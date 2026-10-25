# Changelog

## [1.5.5](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.5.4...arr/sonarr/v1.5.5) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.6.0 ([#158](https://github.com/golusoris/goenvoy/issues/158)) ([241bfd7](https://github.com/golusoris/goenvoy/commit/241bfd7f6f13ec41ac18b28c73776ae40d368a7a))

## [1.5.4](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.5.3...arr/sonarr/v1.5.4) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.5.3](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.5.2...arr/sonarr/v1.5.3) (2026-10-19)


### Code Refactoring

* **arr:** build the service clients through arr.NewServiceClient ([#137](https://github.com/golusoris/goenvoy/issues/137)) ([f3b7f95](https://github.com/golusoris/goenvoy/commit/f3b7f952b56fc34584869e2ef6ab12e25d843a9a))

## [1.5.2](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.5.1...arr/sonarr/v1.5.2) (2026-10-19)

### Code Refactoring

* **arr:** fold log, filesystem-type and manual-import methods onto arr/v2 helpers ([#132](https://github.com/golusoris/goenvoy/issues/132)) ([76166fb](https://github.com/golusoris/goenvoy/commit/76166fb87491be769473e2eb8eaa71860eba554f))

## [1.5.1](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.5.0...arr/sonarr/v1.5.1) (2026-10-18)

### Code Refactoring

* **arr:** fold sonarr, radarr and whisparr onto the arr/v2 resource helpers ([#123](https://github.com/golusoris/goenvoy/issues/123)) ([2e68758](https://github.com/golusoris/goenvoy/commit/2e68758f8ab1802345882005129caf44ab748ed1))
* **arr:** fold the request tail of the remaining arr methods onto the arr/v2 helpers ([#126](https://github.com/golusoris/goenvoy/issues/126)) ([21144ce](https://github.com/golusoris/goenvoy/commit/21144ce9e343c81b8ce667cdc6378fcace46ba91))

## [1.5.0](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.4.0...arr/sonarr/v1.5.0) (2026-10-18)

### Features

* **arr:** add typed manual import candidates and the ManualImport command to sonarr and radarr ([#112](https://github.com/golusoris/goenvoy/issues/112)) ([e77d452](https://github.com/golusoris/goenvoy/commit/e77d45263f3b8406c6cc01f206d60a7331293327))

## [1.4.0](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.3.1...arr/sonarr/v1.4.0) (2026-10-17)

### Features

* **arr:** add queue options to sonarr, radarr, lidarr and readarr ([#110](https://github.com/golusoris/goenvoy/issues/110)) ([42f31c1](https://github.com/golusoris/goenvoy/commit/42f31c115d9afc057f6433fb25dd991924360a72))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.3.0...arr/sonarr/v1.3.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.1.0 ([#64](https://github.com/golusoris/goenvoy/issues/64)) ([273caf4](https://github.com/golusoris/goenvoy/commit/273caf4d61faa26b1d2b9f79e7ca4539025a5f35))

### Code Refactoring

* **arr:** split 6 monolith files into per-resource files ([2264722](https://github.com/golusoris/goenvoy/commit/2264722d503b9f2c9788f94bf99d3ec49e4f25d0))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/sonarr/v1.2.0...arr/sonarr/v1.3.0) (2026-04-15)

### Features

* add HeadPing, UploadBackup, GetRaw to all arr packages ([893a185](https://github.com/golusoris/goenvoy/commit/893a185c977a366421599d7d738f15f103151c2c))
* **arr/sonarr:** add Sonarr v3/v4 API client ([bfb903b](https://github.com/golusoris/goenvoy/commit/bfb903b1c87f94f18c5cbac2e211b012ea636f29))

### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
