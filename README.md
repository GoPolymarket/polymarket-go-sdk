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

## 📈 Polymarket趋势与SDK定位

- **链上预测市场走向机构化**：合规团队与机构交易需要标准化 SDK，统一签名、风控与连接管理。
- **实时数据与事件驱动**：CLOB 与 WebSocket 订阅成为策略核心，SDK 提供低延迟的订阅与心跳管理能力。
- **交易基础设施走向安全合规**：企业级密钥管理、审计与安全扫描成为默认配置，本 SDK 以 KMS 与安全审计文档为核心支撑。

## 🏗 Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a deep dive into the modular design and technical roadmap.

```text
pkg/
├── auth/              # Auth & Signing (EOA, AWS KMS)
│   ├── kms/           # AWS KMS Integration (EIP-712)
│   └── ...
├── clob/              # CLOB REST Core
```

## 🔐 Security & Compliance

See [docs/SECURITY.md](docs/SECURITY.md) for details on AWS KMS integration and the security model of the remote builder signer.

A full security audit checklist and CI guidance are captured in [docs/SECURITY_AUDIT.md](docs/SECURITY_AUDIT.md).

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

## ✅ 使用场景

- **量化做市与套利**：统一的订单与行情接口，方便搭建跨市场策略。
- **机构风控交易**：KMS 与审计流程确保密钥与访问控制合规。
- **实时风控/预警**：WebSocket 与 RTDS 组合实现实时监控与风控信号。
- **研究与数据分析**：统一 API 结构便于数据拉取与事件回测。

## 🗺 技术路线与Roadmap

- [x] Full CLOB REST Support
- [x] Modular RFQ & WebSocket subsystems
- [x] **AWS KMS Integration**
- [x] Security audit documentation + CI vulnerability scan
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
