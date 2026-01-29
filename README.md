# Polymarket Go SDK

[![Go CI](https://github.com/GoPolymarket/polymarket-go-sdk/actions/workflows/go.yml/badge.svg)](https://github.com/GoPolymarket/polymarket-go-sdk/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/GoPolymarket/polymarket-go-sdk.svg)](https://pkg.go.dev/github.com/GoPolymarket/polymarket-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An unofficial, production-ready, and feature-complete Go SDK for the Polymarket CLOB (Central Limit Order Book). Designed for high-frequency trading, market making, and data analysis.

> **Note**: This is a community-maintained project and is not officially affiliated with Polymarket. We aim to provide a high-quality, spec-compliant implementation that can be relied upon by professionals.

## ✨ Key Features

- **🛡️ Type-Safe & Robust**: Fully typed API responses and requests using standard Go patterns.
- **🚀 High Performance**: Optimized for low-latency trading with efficient HTTP and WebSocket handling.
- **🔌 WebSocket Recovery**: Built-in "dead connection" detection and automatic reconnection logic (heartbeat monitoring).
- **🔐 Secure Authentication**:
    - Supports L1 (EIP-712) and L2 (HMAC) signing.
    - **Proxy Wallet & Gnosis Safe**: Deterministic address derivation for smart contract wallets.
    - **AWS KMS**: Plug-and-play support for enterprise-grade key management.
- **🧩 Modular Architecture**: Clean separation of concerns (CLOB, Account, Market Data, WebSocket).
- **🏗️ Builder Attribution**: Easy configuration for builder rewards attribution.

## 🏗 Architecture

The SDK is organized into modular packages to ensure maintainability and extensibility:

```mermaid
graph TD
    Client[Client Interface] --> Transport[Transport Layer]
    Client --> CLOB[CLOB Module]
    Client --> Auth[Auth Module]
    Client --> WS[WebSocket Module]
    
    CLOB --> Orders[Order Management]
    CLOB --> Markets[Market Data]
    CLOB --> Account[Account & Rewards]
    
    Auth --> EOA[Private Key Signer]
    Auth --> KMS[AWS KMS Signer]
    
    WS --> Stream[Event Streams]
```

- **`pkg/clob`**: The core client for REST API interactions (Orders, Markets, Account).
- **`pkg/clob/ws`**: Robust WebSocket client with auto-reconnect and typed event channels.
- **`pkg/auth`**: Cryptographic primitives for EIP-712 signing and HMAC generation.
- **`pkg/transport`**: HTTP transport layer handling signing injection, retries, and error parsing.

## 🚀 Installation

```bash
go get github.com/GoPolymarket/polymarket-go-sdk
```

## 🛠 Quick Start

### 1. Initialize Client & Authentication

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func main() {
	// 1. Initialize Signer (Private Key or KMS)
	pk := os.Getenv("POLYMARKET_PK")
	signer, err := auth.NewPrivateKeySigner(pk, 137) // 137 = Polygon Mainnet
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize Credentials
	creds := &auth.APIKey{
		Key:        os.Getenv("POLY_API_KEY"),
		Secret:     os.Getenv("POLY_API_SECRET"),
		Passphrase: os.Getenv("POLY_API_PASSPHRASE"),
	}

	// 3. Create Client
	client := polymarket.NewClient(
		polymarket.WithAuth(signer, creds),
	)

	// 4. Check System Status
	status, _ := client.CLOB().Health(context.Background())
	fmt.Println("System Status:", status)
}
```

### 2. Place an Order (Complex Signing Made Easy)

The SDK handles the complex EIP-712 hashing and signing automatically.

```go
ctx := context.Background()

// Create an order builder
resp, err := client.CLOB().CreateOrder(ctx, 
    clobtypes.NewOrderBuilder(client.CLOB(), signer).
        TokenID("TOKEN_ID_HERE").
        Side("BUY").
        Price(0.50).
        Size(100.0).
        OrderType(clobtypes.OrderTypeGTC).
        Build(),
)

if err != nil {
    log.Fatal("Order failed:", err)
}
fmt.Printf("Order Placed: %s\n", resp.ID)
```

### 3. Stream Market Data (WebSocket)

```go
wsClient := client.CLOB().WS()
defers wsClient.Close()

// Subscribe to price updates
sub, err := wsClient.SubscribePrices(ctx, []string{"TOKEN_ID_HERE"})
if err != nil {
    log.Fatal(err)
}

for event := range sub {
    fmt.Printf("New Price for %s: %s\n", event.AssetID, event.Price)
}
```

### 4. Fetch All Markets (Auto-Pagination)

Forget about manually handling `next_cursor`.

```go
// Automatically iterates through all pages
allMarkets, err := client.CLOB().MarketsAll(ctx, &clobtypes.MarketsRequest{
    Active: boolPtr(true),
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Fetched %d active markets\n", len(allMarkets))
```

## 🗺 Roadmap

We are committed to maintaining this SDK as the best-in-class solution for Polymarket.

- [x] **Core CLOB REST API**: Complete coverage of Order, Market, and Account endpoints.
- [x] **WebSocket Client**: Robust, auto-reconnecting stream client.
- [x] **Authentication**: Support for EOA, Proxy, Safe, and AWS KMS.
- [x] **Pagination**: Helper methods for automatic resource iteration.
- [x] **CI/CD**: Linting, Testing, and Coverage enforcement.
- [ ] **Gamma API**: Full integration with the Gamma (Market Metadata) API.
- [ ] **CTF Exchange**: Direct interaction with the CTF Exchange contract.
- [ ] **CLI Tool**: A standalone CLI for managing orders and keys.

## 🤝 Contributing

Contributions are welcome! Please check out the `examples/` directory for more usage patterns.

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'Add some amazing feature'`).
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a Pull Request.

**Requirements**:
*   Pass `go test ./...`
*   Pass `golangci-lint run ./...`
*   Maintain >10% test coverage (strictly enforced by CI).

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.

---

*This project is an independent effort to provide a high-quality Go ecosystem for Polymarket. If you find it useful, please star the repo!*