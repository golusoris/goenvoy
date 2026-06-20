# Changelog

## Unreleased

- **BREAKING:** `New` now returns `(*Client, error)` and validates the base URL. Callers must check the error.

## [2.0.0](https://github.com/golusoris/goenvoy/compare/downloadclient/deluge/v1.3.0...downloadclient/deluge/v2.0.0) (2026-06-19)


### ⚠ BREAKING CHANGES

* client import paths gain a major-version suffix.
* constructor sweep — 23 modules return (*Client, error)

### Features

* bump module major versions for 22 clients with breaking constructors ([7309996](https://github.com/golusoris/goenvoy/commit/73099965d27223a3ebd5d65a0cf61b1ac21770c0))
* constructor sweep — 23 modules return (*Client, error) ([2d37053](https://github.com/golusoris/goenvoy/commit/2d37053ff32101d4bd51aaf0247a87fc22b7368f))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/downloadclient/deluge/v1.2.0...downloadclient/deluge/v1.3.0) (2026-04-15)


### Features

* add download client modules (qBittorrent, Transmission, Deluge, SABnzbd, NZBGet, rTorrent) ([51326c8](https://github.com/golusoris/goenvoy/commit/51326c876638f6fb30f6b27b4a4c78726e6c4f98))


### Bug Fixes

* remove dead Version struct in deluge, add json tag to anilist APIError ([d992a4b](https://github.com/golusoris/goenvoy/commit/d992a4bd3c4e79e817b6fbf21a8b84ba8f916168))
