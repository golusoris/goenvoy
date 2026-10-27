# Changelog

## [1.1.3](https://github.com/golusoris/goenvoy/compare/downloadclient/queue/v1.1.2...downloadclient/queue/v1.1.3) (2026-10-27)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/downloadclient/deluge/v2 to v2.0.3 ([#189](https://github.com/golusoris/goenvoy/issues/189)) ([edc8fed](https://github.com/golusoris/goenvoy/commit/edc8fed83deaae8c93dc1ab3cdc8e152ecbeff6c))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/nzbget/v2 to v2.0.3 ([#190](https://github.com/golusoris/goenvoy/issues/190)) ([a0e5ba7](https://github.com/golusoris/goenvoy/commit/a0e5ba70857f35a1b053b00dc53767bffea1bc76))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/qbit/v2 to v2.0.3 ([#191](https://github.com/golusoris/goenvoy/issues/191)) ([b575359](https://github.com/golusoris/goenvoy/commit/b5753599b4da313b172006f665c57282622d1fad))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/rtorrent/v2 to v2.1.2 ([#196](https://github.com/golusoris/goenvoy/issues/196)) ([fc8fc78](https://github.com/golusoris/goenvoy/commit/fc8fc783930ca78573caf1855da1ab5ec99c8e6c))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/sabnzbd/v2 to v2.0.3 ([#192](https://github.com/golusoris/goenvoy/issues/192)) ([8aa8f08](https://github.com/golusoris/goenvoy/commit/8aa8f089f8863f7a4ea3cf27fd40cd4fc0ee0b6a))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/transmission/v2 to v2.0.3 ([#193](https://github.com/golusoris/goenvoy/issues/193)) ([fc82306](https://github.com/golusoris/goenvoy/commit/fc82306bacf2994bfff51e23a096a9b5a187927a))

## [1.1.2](https://github.com/golusoris/goenvoy/compare/downloadclient/queue/v1.1.1...downloadclient/queue/v1.1.2) (2026-10-26)


### Bug Fixes

* give every default HTTP client a private transport pool ([#183](https://github.com/golusoris/goenvoy/issues/183)) ([367d63f](https://github.com/golusoris/goenvoy/commit/367d63f8cecc2de3c4fc98a1e3d3f69c11c5080c))

## [1.1.1](https://github.com/golusoris/goenvoy/compare/downloadclient/queue/v1.1.0...downloadclient/queue/v1.1.1) (2026-10-25)


### Bug Fixes

* **deps:** update module github.com/golusoris/goenvoy/downloadclient/deluge/v2 to v2.0.1 ([#163](https://github.com/golusoris/goenvoy/issues/163)) ([3a53ebf](https://github.com/golusoris/goenvoy/commit/3a53ebf3c8df1a7ac52b9c8694b21e94ce9b3067))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/nzbget/v2 to v2.0.1 ([#164](https://github.com/golusoris/goenvoy/issues/164)) ([f987529](https://github.com/golusoris/goenvoy/commit/f9875296dd16e941ebf3cb4144fe306a1ebb7a4b))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/qbit/v2 to v2.0.1 ([#165](https://github.com/golusoris/goenvoy/issues/165)) ([57ba6ab](https://github.com/golusoris/goenvoy/commit/57ba6ab7704b1eb20ff7b9f790a9f6d27fb0781e))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/sabnzbd/v2 to v2.0.1 ([#180](https://github.com/golusoris/goenvoy/issues/180)) ([2e560b8](https://github.com/golusoris/goenvoy/commit/2e560b83e174dacecc76e26983f69d1106ab2a6e))
* **deps:** update module github.com/golusoris/goenvoy/downloadclient/transmission/v2 to v2.0.1 ([#181](https://github.com/golusoris/goenvoy/issues/181)) ([89505a9](https://github.com/golusoris/goenvoy/commit/89505a9449502019c35b41f630a93239f7d1706a))

## [1.1.0](https://github.com/golusoris/goenvoy/compare/downloadclient/queue/v1.0.0...downloadclient/queue/v1.1.0) (2026-10-23)


### Features

* **downloadclient/rtorrent:** expose the hash queue state as Torrent.Hashing ([#167](https://github.com/golusoris/goenvoy/issues/167)) ([7764872](https://github.com/golusoris/goenvoy/commit/7764872073a4c10bc4b618362d8e28e832bc4d6c))


### Bug Fixes

* **deps:** move to Go 1.26.10 for the crypto/x509 advisory GO-2026-6809 ([#159](https://github.com/golusoris/goenvoy/issues/159)) ([baceba1](https://github.com/golusoris/goenvoy/commit/baceba1d84672942b8ce2e9ba0cf9b879337fac1))

## 1.0.0 (2026-10-17)

### Features

* **downloadclient/queue:** add a unified read-only queue snapshot ([#100](https://github.com/golusoris/goenvoy/issues/100)) ([9dbe5a3](https://github.com/golusoris/goenvoy/commit/9dbe5a3961e13c99da1602a2501de57d75cc2fb7))
