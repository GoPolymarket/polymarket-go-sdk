# Security & Compliance

## Security Posture Overview

- **Key management isolation**: Signing keys never leave KMS or your HSM boundary.
- **Least-privilege access**: SDK usage is scoped via API keys and IAM policies.
- **Transport hardening**: All network access uses TLS and centralized transport configuration.
- **Auditability**: Designed for CloudTrail and centralized logging integration.

## AWS KMS Integration

This SDK provides native support for **AWS KMS (Key Management Service)**, allowing institutions to sign EIP-712 orders without exporting private keys.

### Why is this important?
- **Non-Custodial**: Your backend servers never store or see the private key.
- **Audit Logs**: AWS CloudTrail logs every signing attempt.
- **Policy Control**: You can restrict which IAM roles can invoke the `Sign` operation.

### How it works
The SDK handles the low-level cryptography required to convert AWS KMS's `ECDSA_SHA_256` output (ASN.1 encoded) into the Ethereum-standard 65-byte `[R || S || V]` format.

```go
import "github.com/GoPolymarket/polymarket-go-sdk/pkg/auth/kms"

// 1. Initialize AWS KMS Client
cfg, _ := config.LoadDefaultConfig(ctx)
kmsClient := kms.NewFromConfig(cfg)

// 2. Create the Signer
// The SDK automatically fetches the public key to verify the signature recovery ID (V).
signer, _ := kms.NewAWSSigner(ctx, kmsClient, "alias/my-trading-key", 137)

// 3. Use it
client := polymarket.NewClient().WithAuth(signer, apiKey)
```

## Builder Attribution Security

By default, this SDK attributes trading volume to the maintainer via a **Remote Signer**.

### Is my API Key safe?
**YES.**
1. **No Credentials Shared**: Your L2 CLOB API Keys (`POLY_API_KEY`) and L1 Private Keys are **NEVER** sent to the remote signer.
2. **Attribution Only**: The remote signer only signs the `POLY_BUILDER_SIGNATURE` header. This header is used solely for volume tracking and cannot be used to withdraw funds or place orders on your behalf.
3. **Transparency**: The source code for the signer service is available in `cmd/signer-server`.

### Opting Out
Institutions can easily override this by providing their own Builder credentials:

```go
client := polymarket.NewClient(
    polymarket.WithBuilderAttribution("YOUR_KEY", "YOUR_SECRET", "YOUR_PASSPHRASE"),
)
```

## Security Operations Checklist

- Rotate API keys and KMS policies regularly.
- Enable CloudTrail alerts for unusual signing activity.
- Use environment-scoped API keys per workload.
- Review CI security scan output (see [docs/SECURITY_AUDIT.md](SECURITY_AUDIT.md)).
