# Changelog

## Unreleased

- Use a private HTTP transport pool for the default client, so `CloseIdleConnections` on `http.DefaultTransport` (for example from `httptest.Server.Close`) no longer breaks in-flight requests.

## [1.1.0](https://github.com/golusoris/goenvoy/compare/downloadclient/queue/v1.0.0...downloadclient/queue/v1.1.0) (2026-10-23)


### Features

* **downloadclient/rtorrent:** expose the hash queue state as Torrent.Hashing ([#167](https://github.com/golusoris/goenvoy/issues/167)) ([7764872](https://github.com/golusoris/goenvoy/commit/7764872073a4c10bc4b618362d8e28e832bc4d6c))


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## 1.0.0 (2026-10-17)

### Features

* **downloadclient/queue:** add a unified read-only queue snapshot ([#100](https://github.com/golusoris/goenvoy/issues/100)) ([9dbe5a3](https://github.com/golusoris/goenvoy/commit/9dbe5a3961e13c99da1602a2501de57d75cc2fb7))
