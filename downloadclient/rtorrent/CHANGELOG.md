# Changelog

## [2.1.4](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.1.3...downloadclient/rtorrent/v2.1.4) (2026-11-12)


### Bug Fixes

* **deps:** move to Go 1.26.11 for the net/http advisory GO-2026-6859 ([#310](https://github.com/golusoris/goenvoy/issues/310)) ([eee19f2](https://github.com/golusoris/goenvoy/commit/eee19f20da04468690d04e67ed870cf85506218e))

## [2.1.3](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.1.2...downloadclient/rtorrent/v2.1.3) (2026-11-08)


### Bug Fixes

* handle error returns that were assigned to blank ([#278](https://github.com/golusoris/goenvoy/issues/278)) ([aa1f0dc](https://github.com/golusoris/goenvoy/commit/aa1f0dce6bf56f7fe65dcdcd782e8fa20a6f3e29))

## [2.1.2](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.1.1...downloadclient/rtorrent/v2.1.2) (2026-10-26)


### Bug Fixes

* give every default HTTP client a private transport pool ([#183](https://github.com/golusoris/goenvoy/issues/183)) ([367d63f](https://github.com/golusoris/goenvoy/commit/367d63f8cecc2de3c4fc98a1e3d3f69c11c5080c))

## [2.1.1](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.1.0...downloadclient/rtorrent/v2.1.1) (2026-10-23)


### Bug Fixes

* **downloadclient/rtorrent:** report short d.multicall2 rows instead of dropping them ([#171](https://github.com/golusoris/goenvoy/issues/171)) ([4ad2ba7](https://github.com/golusoris/goenvoy/commit/4ad2ba752a3c68bc8ff17ad49f3c92ba7b874b72))

## [2.1.0](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.0.1...downloadclient/rtorrent/v2.1.0) (2026-10-23)


### Features

* **downloadclient/rtorrent:** expose the hash queue state as Torrent.Hashing ([#167](https://github.com/golusoris/goenvoy/issues/167)) ([7764872](https://github.com/golusoris/goenvoy/commit/7764872073a4c10bc4b618362d8e28e832bc4d6c))


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [2.0.1](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v2.0.0...downloadclient/rtorrent/v2.0.1) (2026-10-11)

### Bug Fixes

- **downloadclient/rtorrent:** read rTorrent's i8 flags and fault struct ([#91](https://github.com/golusoris/goenvoy/issues/91)) ([ecccd62](https://github.com/golusoris/goenvoy/commit/ecccd627f210023ef81a409dbca18b5a04bcba6b)), closes [#89](https://github.com/golusoris/goenvoy/issues/89)

## [2.0.0](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v1.3.0...downloadclient/rtorrent/v2.0.0) (2026-06-19)

### ⚠ BREAKING CHANGES

- client import paths gain a major-version suffix.
- constructor sweep — 23 modules return (*Client, error)

### Features

- bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
- constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/downloadclient/rtorrent/v1.2.0...downloadclient/rtorrent/v1.3.0) (2026-04-15)

### Features

- add download client modules (qBittorrent, Transmission, Deluge, SABnzbd, NZBGet, rTorrent) ([51326c8](https://github.com/golusoris/goenvoy/commit/51326c876638f6fb30f6b27b4a4c78726e6c4f98))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- **BREAKING:** `New` now returns `(*Client, error)` and validates the endpoint URL. Callers must check the error.
