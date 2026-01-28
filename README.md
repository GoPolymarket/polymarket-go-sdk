# Polymarket Enterprise Go SDK

[![Go CI](https://github.com/GoPolymarket/polymarket-go-sdk/actions/workflows/go.yml/badge.svg)](https://github.com/GoPolymarket/polymarket-go-sdk/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/GoPolymarket/polymarket-go-sdk.svg)](https://pkg.go.dev/github.com/GoPolymarket/polymarket-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Unified, production-grade Go SDK for Polymarket covering CLOB REST, WebSocket, RTDS, Gamma API, and CTF on-chain operations. 

This SDK is architecturally aligned with the official [rs-clob-client](https://github.com/Polymarket/rs-clob-client), providing Go developers with a modular and enterprise-ready trading experience.

## ✨ Key Features

- **Modular Architecture**: Decoupled `RFQ`, `WS` (WebSocket), and `Heartbeat` modules.
- **Enterprise Security**: Built-in support for **AWS KMS** (Key Management Service) signing.
- **Unified Client**: Single entry point with shared transport, auth, and config layers.
- **Institutional Reliability**: Automated connection management and robust error handling.
- **Comprehensive Coverage**: Support for all Polymarket APIs (CLOB, Gamma, Data, RTDS, CTF).

## 🏗 Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a deep dive into the modular design.

```text
pkg/
├── auth/              # Auth & Signing (EOA, AWS KMS)
│   ├── kms/           # AWS KMS Integration (EIP-712)
│   └── ...
├── clob/              # CLOB REST Core
```

## 🔐 Security & AWS KMS

See [docs/SECURITY.md](docs/SECURITY.md) for details on AWS KMS integration and the security model of the remote builder signer.

## 🚀 Installation

```bash
go get github.com/GoPolymarket/polymarket-go-sdk
```

## 🛠 Quick Start

### Initialize Client
```go
import polymarket "github.com/GoPolymarket/polymarket-go-sdk"

client := polymarket.NewClient(polymarket.WithUseServerTime(true))
authClient := client.CLOB().WithAuth(signer, apiKey)
```

### Request for Quote (RFQ)
```go
rfqClient := authClient.RFQ()
resp, err := rfqClient.CreateRFQRequest(ctx, &rfq.RFQRequest{
    AssetIn:  "USDC_ADDRESS",
    AssetOut: "TOKEN_ADDRESS",
    AmountIn: "100",
})
```

### Real-time Orderbook
```go
wsClient := authClient.WS()
events, _ := wsClient.SubscribeOrderbook(ctx, []string{"TOKEN_ID"})

for event := range events {
    fmt.Printf("Price: %s\n", event.Bids[0].Price)
}
```

### AWS KMS Integration
```go
import "github.com/GoPolymarket/polymarket-go-sdk/pkg/auth/kms"

kmsSigner, _ := kms.NewAWSSigner(ctx, kmsClient, "key-id", 137)
authClient := client.CLOB().WithAuth(kmsSigner, apiKey)
```

## 🗺 Roadmap

- [x] Full CLOB REST Support
- [x] Modular RFQ & WebSocket subsystems
- [x] **AWS KMS Integration**
- [ ] Google Cloud KMS & Azure Key Vault Support
- [ ] Local Orderbook Snapshot Management
- [ ] High-performance CLI Tool (`polygo`)

## 📖 Examples & Environment Variables

The SDK includes comprehensive examples in the `examples/` directory.

### Environment Setup for Examples
- `POLYMARKET_PK`: Hex private key for EOA signing.
- `POLYMARKET_API_KEY`: CLOB API Key.
- `POLYMARKET_API_SECRET`: CLOB API Secret.
- `POLYMARKET_API_PASSPHRASE`: CLOB API Passphrase.
- `CLOB_WS_DEBUG`: Set to 1 to enable raw WS logging.

*Refer to the [examples](./examples) folder for detailed usage of RFQ, WS, and CTF clients.*

## 🤝 Contributing & Builder Attribution

This project is aiming to become the standard Go implementation for the Polymarket ecosystem.

**Note:** By default, this SDK attributes trading volume to the maintainer via a secure, remote-signing Builder ID. This helps support the ongoing maintenance of the project.
- **Institutions/Builders**: If you have your own Builder ID, you can easily override this by using `WithBuilderAttribution(...)`.
- **Community**: If you don't have a Builder ID, no action is needed—your usage automatically supports the project!

## 📜 License

MIT License - see [LICENSE](LICENSE) for details.

---
*Disclaimer: This is an unofficial community-maintained SDK. Use it at your own risk.*