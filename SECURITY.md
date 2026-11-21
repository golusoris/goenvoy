# Security policy

## Reporting a vulnerability

Please **do not** open public GitHub issues for security vulnerabilities.

Email: `security@lusoris.dev` (or open a private security advisory on GitHub).

We aim to acknowledge within 72 hours and provide a remediation plan within 7 days.

## Supported versions

Only the latest release of each module is patched. Downstream apps should bump promptly when an advisory is published for a module they use.

## Supply chain

Every tagged module release ships:

- Source tarball `<module>_<version>_source.tar.gz` and its `<module>_<version>_source.checksums.txt`, where `<module>` is the module path with `/` replaced by `_` (for example `metadata_video_tvmaze_v1.3.7_source.tar.gz`).
- `<asset>.sigstore.json` for every asset — [cosign](https://docs.sigstore.dev/cosign/) keyless Sigstore bundle (GitHub OIDC) holding the signature, certificate and transparency-log proof.
- SPDX SBOM via [syft](https://github.com/anchore/syft).
- SLSA-L3 provenance via [actions/attest-build-provenance](https://github.com/actions/attest-build-provenance).

Verify a release archive:

```bash
cosign verify-blob \
  --bundle <module>_<version>_source.checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/(golusoris|plurano)/goenvoy/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  <module>_<version>_source.checksums.txt
sha256sum -c <module>_<version>_source.checksums.txt
```

## Security practices

- **Pure stdlib** — zero external dependencies reduces supply-chain risk to the Go toolchain and `golang.org/x/*` (none currently used).
- **No telemetry** — clients open connections only to the service the caller points them at.
- **No secret persistence** — API keys, bearer tokens, and refresh tokens live in the `*Client` struct only. They are never cached to disk, never written to logs, never included in error messages.
- **TLS on by default** — no module disables certificate verification. `WithHTTPClient` lets callers override the `*http.Client` for specialised cases; doing so is their responsibility.
- **URL validation** — every `New` validates the `baseURL` scheme (`http`/`https` only) and parseability.
- **Static analysis** — every module passes [gosec](https://github.com/securego/gosec) + [golangci-lint](https://golangci-lint.run/) (incl. `errorlint`, `noctx`, `bodyclose`, `contextcheck`, `containedctx`) + [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) + CodeQL (`security-extended,security-and-quality`).
- **OSSF Scorecard** — published and monitored; see the badge on `README.md`.

## Dependencies

Tracked by [Dependabot](https://docs.github.com/en/code-security/dependabot):

- `gomod` — pure-stdlib means `gomod` is quiet; updates only touch test-only tooling pulled via `go install`, if any.
- `github-actions` — all workflow actions are hash-pinned; Dependabot proposes version bumps with updated hashes.

## Scope

This library is a collection of HTTP API clients. Relevant security concerns include:

- Credential leakage (API keys / tokens in logs or error messages).
- Request injection (path traversal, header injection).
- TLS verification bypass.
- Improper error handling exposing sensitive data.

We actively test against these in code review + `gosec` rules G101, G107, G306, G402, G505.
