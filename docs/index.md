# goenvoy

goenvoy is a collection of pure-stdlib Go client libraries for media
automation, metadata, and download service APIs — 70 modules covering 63+
services in a multi-module monorepo with zero external dependencies. Each
service lives in its own independently versioned Go module so applications
import only what they need.

The name **goenvoy** combines **Go** and **envoy** — a messenger sent to deal
with another party on someone's behalf, which is the role each module plays for
one service API. It is pronounced "go-EN-voy" (/ɡoʊ ˈɛnvɔɪ/).

## Install

Install any service client module directly with `go get`:

```bash
go get github.com/golusoris/goenvoy/arr/sonarr
```

Most client constructors follow `New(baseURL, apiKey, opts...)`. Some differ:
`kitsu.New` takes options only, and `nzbget.New` takes a base URL, username, and
password. See [ADR-0003](adr/0003-new-constructor-shape.md) and the
[package reference](https://pkg.go.dev/github.com/golusoris/goenvoy) for exact
signatures. I/O methods take `context.Context` as their first argument.

## Documentation map

- **[Principles](principles.md)** — coding rules (Power of 10), SEI CERT Go
  practices, Go style, supply-chain controls, wire protocols, and testing
  standards.
- **[Triage runbook](triage-runbook.md)** — maintainer workflow for reviewing
  issues, pull requests, and automated checks.
- **[Architecture Decisions](adr/README.md)** — Nygard-format records for
  module boundaries, constructors, functional options, error types, and OAuth2
  flows.
- **[Upstream API notes](https://github.com/golusoris/goenvoy/tree/main/docs/upstream)**
  — per-service upstream API documentation URLs, target versions, and
  verification dates, grouped by category in the site navigation (for example,
  [Sonarr](upstream/arr-sonarr.md)).
- **[Go package reference](https://pkg.go.dev/github.com/golusoris/goenvoy)**
  — generated API documentation and runnable examples on `pkg.go.dev`.
- **[Contributing](https://github.com/golusoris/goenvoy/blob/main/CONTRIBUTING.md)**
  — Conventional Commits, CI gates, local development commands, and release
  workflow.

The source lives on [GitHub](https://github.com/golusoris/goenvoy).
