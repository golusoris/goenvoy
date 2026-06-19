# Changelog

## [1.3.1](https://github.com/golusoris/goenvoy/compare/arr/whisparr/v1.3.0...arr/whisparr/v1.3.1) (2026-06-19)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/arr/v2 to v2.1.0 ([#64](https://github.com/golusoris/goenvoy/issues/64)) ([273caf4](https://github.com/golusoris/goenvoy/commit/273caf4d61faa26b1d2b9f79e7ca4539025a5f35))


### Code Refactoring

* **arr:** split 6 monolith files into per-resource files ([2264722](https://github.com/golusoris/goenvoy/commit/2264722d503b9f2c9788f94bf99d3ec49e4f25d0))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/arr/whisparr/v1.2.0...arr/whisparr/v1.3.0) (2026-04-15)


### Features

* add HeadPing, UploadBackup, GetRaw to all arr packages ([893a185](https://github.com/golusoris/goenvoy/commit/893a185c977a366421599d7d738f15f103151c2c))
* **arr/whisparr:** add Whisparr v2 and v3/Eros API clients ([6c47600](https://github.com/golusoris/goenvoy/commit/6c476003e0a83e1776728608014b5d594c896d3c))


### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))


### Code Refactoring

* rename ErosClient to ClientV3, remove Eros codename from public API ([d7028b5](https://github.com/golusoris/goenvoy/commit/d7028b5a179279bd7190033b17ae9a8a2978e542))
