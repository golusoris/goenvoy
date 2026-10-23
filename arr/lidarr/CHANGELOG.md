# Changelog

## [1.6.1](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.6.0...arr/lidarr/v1.6.1) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.6.0](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.5.3...arr/lidarr/v1.6.0) (2026-10-19)


### Features

* **arr/lidarr:** add typed manual import candidates, command and rescore ([#156](https://github.com/golusoris/goenvoy/issues/156)) ([aa66298](https://github.com/golusoris/goenvoy/commit/aa662988d9b1d2fc31667e0b14fae38d2c3b52d6))
* **arr/readarr:** add typed manual import candidates and command ([#156](https://github.com/golusoris/goenvoy/issues/156)) ([aa66298](https://github.com/golusoris/goenvoy/commit/aa662988d9b1d2fc31667e0b14fae38d2c3b52d6))
* **arr/whisparr:** add queue options and delete options for v2 and v3 ([#156](https://github.com/golusoris/goenvoy/issues/156)) ([aa66298](https://github.com/golusoris/goenvoy/commit/aa662988d9b1d2fc31667e0b14fae38d2c3b52d6))

## [1.5.3](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.5.2...arr/lidarr/v1.5.3) (2026-10-19)


### Code Refactoring

* **arr:** build the service clients through arr.NewServiceClient ([#137](https://github.com/golusoris/goenvoy/issues/137)) ([f3b7f95](https://github.com/golusoris/goenvoy/commit/f3b7f952b56fc34584869e2ef6ab12e25d843a9a))

## [1.5.2](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.5.1...arr/lidarr/v1.5.2) (2026-10-19)

### Code Refactoring

* **arr:** fold log, filesystem-type and manual-import methods onto arr/v2 helpers ([#132](https://github.com/golusoris/goenvoy/issues/132)) ([76166fb](https://github.com/golusoris/goenvoy/commit/76166fb87491be769473e2eb8eaa71860eba554f))

## [1.5.1](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.5.0...arr/lidarr/v1.5.1) (2026-10-18)

### Code Refactoring

* **arr:** fold lidarr, readarr and prowlarr onto the arr/v2 resource helpers ([#125](https://github.com/golusoris/goenvoy/issues/125)) ([0e5fc1d](https://github.com/golusoris/goenvoy/commit/0e5fc1dd3290ce665ef730519452831f7ce4207e))
* **arr:** fold the request tail of the remaining arr methods onto the arr/v2 helpers ([#126](https://github.com/golusoris/goenvoy/issues/126)) ([21144ce](https://github.com/golusoris/goenvoy/commit/21144ce9e343c81b8ce667cdc6378fcace46ba91))

## [1.5.0](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.4.0...arr/lidarr/v1.5.0) (2026-10-18)

### Features

* **arr:** add typed manual import candidates and the ManualImport command to sonarr and radarr ([#112](https://github.com/golusoris/goenvoy/issues/112)) ([e77d452](https://github.com/golusoris/goenvoy/commit/e77d45263f3b8406c6cc01f206d60a7331293327))

## [1.4.0](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.3.1...arr/lidarr/v1.4.0) (2026-10-17)

### Features

* **arr:** add queue options to sonarr, radarr, lidarr and readarr ([#110](https://github.com/golusoris/goenvoy/issues/110)) ([42f31c1](https://github.com/golusoris/goenvoy/commit/42f31c115d9afc057f6433fb25dd991924360a72))

## [1.3.1](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.3.0...arr/lidarr/v1.3.1) (2026-06-19)

### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.1.0 ([#64](https://github.com/golusoris/goenvoy/issues/64)) ([273caf4](https://github.com/golusoris/goenvoy/commit/273caf4d61faa26b1d2b9f79e7ca4539025a5f35))

### Code Refactoring

* **arr:** split 6 monolith files into per-resource files ([2264722](https://github.com/golusoris/goenvoy/commit/2264722d503b9f2c9788f94bf99d3ec49e4f25d0))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/lidarr/v1.2.0...arr/lidarr/v1.3.0) (2026-04-15)

### Features

* add HeadPing, UploadBackup, GetRaw to all arr packages ([893a185](https://github.com/golusoris/goenvoy/commit/893a185c977a366421599d7d738f15f103151c2c))
* **arr/lidarr:** add Lidarr v1 API client ([67431d9](https://github.com/golusoris/goenvoy/commit/67431d93e24bbfdfa9d7adecb8635007af232ff7))

### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
