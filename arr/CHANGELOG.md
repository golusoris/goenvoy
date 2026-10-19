# Changelog

## [2.5.0](https://github.com/golusoris/goenvoy/compare/arr/v2.4.0...arr/v2.5.0) (2026-10-19)


### Features

* **arr:** add NewServiceClient for one-call service constructors ([#136](https://github.com/golusoris/goenvoy/issues/136)) ([0b08cf6](https://github.com/golusoris/goenvoy/commit/0b08cf6f02d3a49e60fad7e49b4ab0126e0d6c03))


### Bug Fixes

* **lint:** hold every Go file to the HISS-04 complexity limits ([#142](https://github.com/golusoris/goenvoy/issues/142)) ([2e5dea9](https://github.com/golusoris/goenvoy/commit/2e5dea9437444f670c5a628a35f8cc11e4883d46)), closes [#76](https://github.com/golusoris/goenvoy/issues/76)

## [2.4.0](https://github.com/golusoris/goenvoy/compare/arr/v2.3.0...arr/v2.4.0) (2026-10-18)


### Features

* **arr:** add GetValue, GetText and SendManualImport helpers ([#130](https://github.com/golusoris/goenvoy/issues/130)) ([a305d72](https://github.com/golusoris/goenvoy/commit/a305d728d03b2334dc30ef73023509f1d70ef310))

## [2.3.0](https://github.com/golusoris/goenvoy/compare/arr/v2.2.0...arr/v2.3.0) (2026-10-18)


### Features

* **arr:** add typed request helpers beside BaseClient ([#121](https://github.com/golusoris/goenvoy/issues/121)) ([788ff61](https://github.com/golusoris/goenvoy/commit/788ff61607deef78945f98a52cded4a6b0fd77d4))

## [2.2.0](https://github.com/golusoris/goenvoy/compare/arr/v2.1.0...arr/v2.2.0) (2026-10-17)


### Features

* **arr/cleanuparr:** add a read-only Cleanuparr client ([#97](https://github.com/golusoris/goenvoy/issues/97)) ([5afd85d](https://github.com/golusoris/goenvoy/commit/5afd85d612c57d1ab46a27e5743c7b059772582c))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/arr/v2.0.0...arr/v2.1.0) (2026-04-15)


### Features

* add HeadPing, UploadBackup, GetRaw to all arr packages ([893a185](https://github.com/golusoris/goenvoy/commit/893a185c977a366421599d7d738f15f103151c2c))
* **arr/bazarr:** add Bazarr API client ([5469769](https://github.com/golusoris/goenvoy/commit/5469769c29c8ab4497cb1c3e2445b572f70eec48))
* **arr/sonarr:** add Sonarr v3/v4 API client ([bfb903b](https://github.com/golusoris/goenvoy/commit/bfb903b1c87f94f18c5cbac2e211b012ea636f29))
* bootstrap monorepo with category base modules ([ffd24fc](https://github.com/golusoris/goenvoy/commit/ffd24fc22b716b41b4a8d27a2ab80950f2c825b7))


### Bug Fixes

* BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))
* use cloned transport in arr.NewBaseClient to prevent cross-test interference ([016baf9](https://github.com/golusoris/goenvoy/commit/016baf97bf8781f8894aa0113ffc134f33c65fc2))
