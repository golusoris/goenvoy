# AGENTS — downloadclient/rtorrent

> Per-service notes. Read [../AGENTS.md](../AGENTS.md) first.

## Upstream API

- Canonical docs: <https://github.com/rakshasa/rtorrent/wiki/rTorrent-0.9-Comprehensive-Command-list-(WIP)>
- Last verified: 2026-04-15
- Details: [docs/upstream/downloadclient-rtorrent.md](../../docs/upstream/downloadclient-rtorrent.md)

## Auth model

Optional HTTP Basic over the XML-RPC (SCGI-via-HTTP) endpoint.

## Wire format

Read from the rTorrent source on 2026-10-11 (v0.16.25, `src/rpc/xmlrpc_tinyxml2.cc`):

- Every integer is an `<i8>`, the 0 and 1 of `d.is_open`, `d.is_active`, `d.complete` and
  `d.is_hash_checking` included. rTorrent writes no `<i4>` and no `<boolean>`; the client
  still reads both.
- A fault is a `<struct>` of `faultCode` (`<i8>`) and `faultString` (`<string>`).

Test fixtures use these tags, so a reader that drops `<i8>` fails the module's tests.
