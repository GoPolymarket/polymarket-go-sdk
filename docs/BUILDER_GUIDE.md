# Polymarket SDK Builder's Guide

Welcome to the Polymarket SDK Builder's Guide! This guide is designed to help you quickly integrate with the Polymarket Builders Program using our SDKs.

The [Polymarket Builders Program](https://github.com/Polymarket/builders-program) rewards developers who build tools, bots, and interfaces that drive volume and liquidity to the Polymarket ecosystem.

## Why use this SDK?

1.  **Automatic Attribution**: Our SDKs are pre-configured to support Builder Attribution headers.
2.  **Simplified Auth**: Handles L1 (Wallet) and L2 (API Key) signatures automatically.
3.  **Type Safety**: Full typing for all CLOB API endpoints.
4.  **Robustness**: Built-in retries and error handling.

---

## How Attribution Works

To receive rewards, every order you place must include specific headers that identify your "Builder API Key".

The SDK supports two modes for Builder Attribution:

1.  **Local Signing (Recommended for independent bots)**: You manage the Builder API Key locally.
2.  **Remote Signing (Recommended for hosted services)**: You use a remote signing service (like the official SDK signer or your own).

### Default Behavior

By default, the SDK uses a **fallback remote signer** attributed to the SDK maintainers. To receive rewards for **your** activity, you must provide your own Builder Configuration.

---

## Quick Start (Go)

### Prerequisites

1.  A Polygon wallet with some MATIC (for gas, though CLOB trading is gasless) and USDC (for trading).
2.  Your Private Key (keep it safe!).

### Step 1: Create a Builder API Key

You can create a Builder API Key programmatically or use an existing one.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    polymarket "github.com/GoPolymarket/polymarket-go-sdk"
    "github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

func main() {
    // 1. Initialize Client & Signer
    pk := os.Getenv("PRIVATE_KEY")
    signer, _ := auth.NewPrivateKeySigner(pk, 137)
    client := polymarket.NewClient(polymarket.WithUseServerTime(true))

    // 2. Derive L2 API Key (Standard Auth)
    l2Client := client.CLOB.WithAuth(signer, nil)
    apiKeyResp, _ := l2Client.DeriveAPIKey(context.Background())
    apiKey := &auth.APIKey{
        Key:        apiKeyResp.APIKey,
        Secret:     apiKeyResp.Secret,
        Passphrase: apiKeyResp.Passphrase,
    }

    // 3. Create Builder API Key
    // IMPORTANT: Save these credentials! You only get them once.
    authClient := client.CLOB.WithAuth(signer, apiKey)
    builderResp, err := authClient.CreateBuilderAPIKey(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Builder Key: %s\n", builderResp.APIKey)
    fmt.Printf("Builder Secret: %s\n", builderResp.Secret)
    fmt.Printf("Builder Passphrase: %s\n", builderResp.Passphrase)
}
```

### Step 2: Configure the Client

Once you have your Builder credentials, configure the client to use them for all orders.

```go
    // ... inside main ...

    // Create a Builder Config with your credentials
    myBuilderConfig := &auth.BuilderConfig{
        Local: &auth.BuilderCredentials{
            Key:        "YOUR_BUILDER_KEY",
            Secret:     "YOUR_BUILDER_SECRET",
            Passphrase: "YOUR_BUILDER_PASSPHRASE",
        },
    }

    // Apply the config to the client
    // All subsequent orders created with 'tradingClient' will be attributed to you.
    tradingClient := authClient.WithBuilderConfig(myBuilderConfig)

    // Example: Create an order
    resp, err := tradingClient.CreateOrder(ctx, &clobtypes.Order{
        TokenID:   "...",
        Price:     0.5,
        Size:      100,
        Side:      clobtypes.Buy,
        OrderType: clobtypes.Limit,
    })
```

---

## Remote Signing (Advanced)

If you are building a platform where users trade via your UI, you might not want to distribute your Builder Private Key to their browsers/clients. Instead, you can set up a remote signer.

```go
    remoteConfig := &auth.BuilderConfig{
        Remote: &auth.BuilderRemoteConfig{
            Host: "https://your-signer-service.com",
            // The SDK will send the order payload to this host to get the builder signature
        },
    }
    
    client := client.WithBuilderConfig(remoteConfig)
```

## Checklist for Builders

- [ ] **Get Sponsored**: Apply for the [Polymarket Builders Program](https://biuilders.polymarket.com).
- [ ] **Generate Key**: Create your unique Builder API Key.
- [ ] **Configure SDK**: Ensure `WithBuilderConfig` is called before placing orders.
- [ ] **Verify**: Place a small test order and check `GET /builder-rewards` (or ask the team) to verify attribution.

## Need Help?

Check out the `examples/builder_flow` directory for a complete, runnable example.
