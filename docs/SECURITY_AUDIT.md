# Security Audit Guide

This document defines the security audit scope for the Polymarket Go SDK and the checks we recommend running locally and in CI.

## Audit Scope

1. **Dependencies & Vulnerabilities**
   - Go module dependency analysis.
   - Vulnerability scanning (CVE matching).
2. **Credential Handling**
   - Ensure secrets are injected via environment variables or secret managers.
   - Avoid hard-coded keys, tokens, or mnemonic phrases.
3. **Signature & Crypto Safety**
   - Confirm KMS signing paths and EIP-712 conversion are used in production.
   - Verify that raw private keys are never logged.
4. **Transport Security**
   - TLS-only communication.
   - Timeouts and retry policies configured by default.
5. **Operational Monitoring**
   - Audit logs enabled for KMS usage.
   - Metrics collection for failed auth and request retries.

## Recommended CI Checks

| Check | Tool | Purpose |
|---|---|---|
| Go build/test | `go test ./...` | Validate correctness and regression coverage. |
| Vulnerability scan | `govulncheck ./...` | Identify known vulnerabilities in dependencies. |

## Running Locally

```bash
go test ./...

go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## CI Integration Notes

The CI workflow runs `govulncheck` by default. If you mirror CI locally, ensure you have access to download modules and the Go toolchain for the same version.
