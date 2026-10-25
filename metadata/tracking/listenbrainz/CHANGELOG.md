# Changelog

## [1.0.3](https://github.com/golusoris/goenvoy/compare/metadata/tracking/listenbrainz/v1.0.2...metadata/tracking/listenbrainz/v1.0.3) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/metadata to v1.4.1 ([#162](https://github.com/golusoris/goenvoy/issues/162)) ([9182fcd](https://github.com/golusoris/goenvoy/commit/9182fcdbb8b8da4c2a248386237c5f7299287094))

## [1.0.2](https://github.com/golusoris/goenvoy/compare/metadata/tracking/listenbrainz/v1.0.1...metadata/tracking/listenbrainz/v1.0.2) (2026-10-23)


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## [1.0.1](https://github.com/golusoris/goenvoy/compare/metadata/tracking/listenbrainz/v1.0.0...metadata/tracking/listenbrainz/v1.0.1) (2026-10-11)

### Bug Fixes

- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([d7c9c63](https://github.com/golusoris/goenvoy/commit/d7c9c635422bda81c3a2f8703e5573a1c62b217d))
- **deps:** update module github.com/golusoris/goenvoy/metadata to v1.3.1 ([#3](https://github.com/golusoris/goenvoy/issues/3)) ([b40adc4](https://github.com/golusoris/goenvoy/commit/b40adc47018b15d3224efa758b39326f7c6f7af1))

## [1.3.0](https://github.com/golusoris/goenvoy/compare/metadata/music/listenbrainz/v1.2.0...metadata/music/listenbrainz/v1.3.0) (2026-04-15)

### Features

- add 8 new service modules (Jackett, NZBHydra2, Spotify, Deezer, IGDB, RAWG, ListenBrainz, Steam) ([2f4f16d](https://github.com/golusoris/goenvoy/commit/2f4f16d34d1b3ffa9ca71d89f9589c2442785a9a))
- **metadata:** add shared BaseClient, migrate all 27 providers, restructure movie→video ([0026baa](https://github.com/golusoris/goenvoy/commit/0026baa54634fa25c00f67d9387e8013e5d70a6e))

### Bug Fixes

- BaseClient.Delete body support, URL injection, default timeouts ([59203c0](https://github.com/golusoris/goenvoy/commit/59203c0eda4c747e89c6567b91811e0cdee46e5b))

## Earlier changes

Recorded by hand before release-please wrote these notes.

- Moved from `metadata/music/listenbrainz` to `metadata/tracking/listenbrainz`. Update imports accordingly.
