# Error Handling Guide

This guide covers how to handle errors effectively when using the Polymarket Go SDK.

## Error Types

The SDK provides structured errors with error codes for programmatic handling:

```go
import (
    "errors"
    sdkerrors "github.com/GoPolymarket/polymarket-go-sdk/pkg/errors"
)
```

### Error Code Categories

| Prefix | Category | Description |
|--------|----------|-------------|
| `AUTH-xxx` | Authentication | Signer, credentials, builder config issues |
| `WALLET-xxx` | Wallet | Proxy/Safe derivation errors |
| `CLOB-xxx` | CLOB API | Trading, orders, markets |
| `NET-xxx` | Network | HTTP, rate limits, circuit breaker |
| `DATA-xxx` | Data API | Market data, positions |
| `WS-xxx` | WebSocket | Subscriptions, connections |
| `CTF-xxx` | CTF | Conditional Token Framework |
| `BRIDGE-xxx` | Bridge | Cross-chain transfers |

## Common Error Handling Patterns

### Authentication Errors

```go
resp, err := client.CLOB.CreateOrder(ctx, order)
if err != nil {
    if errors.Is(err, sdkerrors.ErrMissingSigner) {
        fmt.Println("Please provide a private key signer")
    } else if errors.Is(err, sdkerrors.ErrMissingCreds) {
        fmt.Println("Please provide API credentials")
    } else if errors.Is(err, sdkerrors.ErrUnauthorized) {
        fmt.Println("Invalid credentials, please check your API keys")
    }
}
```

### Trading Errors

```go
resp, err := client.CLOB.CreateOrder(ctx, order)
if err != nil {
    if errors.Is(err, sdkerrors.ErrInsufficientFunds) {
        fmt.Println("Please deposit more USDC to your account")
    } else if errors.Is(err, sdkerrors.ErrRateLimitExceeded) {
        fmt.Println("Backing off due to rate limits...")
        time.Sleep(time.Minute)
    } else if errors.Is(err, sdkerrors.ErrMarketClosed) {
        fmt.Println("This market has been closed")
    } else if errors.Is(err, sdkerrors.ErrGeoblocked) {
        fmt.Println("Trading is not available in your region")
    }
}
```

### Network Errors

```go
resp, err := client.CLOB.GetMarkets(ctx, request)
if err != nil {
    if errors.Is(err, sdkerrors.ErrCircuitOpen) {
        fmt.Println("Service temporarily unavailable, retrying...")
    } else if errors.Is(err, sdkerrors.ErrTooManyRequests) {
        fmt.Println("Too many requests, backing off...")
    }
}
```

## Working with SDK Errors

### Error Code Inspection

```go
resp, err := client.CLOB.CreateOrder(ctx, order)
if err != nil {
    var sdkErr *sdkerrors.SDKError
    if errors.As(err, &sdkErr) {
        fmt.Printf("Error Code: %s\n", sdkErr.Code)
        fmt.Printf("Error Message: %s\n", sdkErr.Message)
    }
}
```

### Custom Error Handling

```go
func handleTradingError(err error) error {
    switch {
    case errors.Is(err, sdkerrors.ErrInsufficientFunds):
        return fmt.Errorf("trading failed: insufficient funds - please deposit USDC")
    case errors.Is(err, sdkerrors.ErrRateLimitExceeded):
        return fmt.Errorf("trading failed: rate limited - please wait before retrying")
    case errors.Is(err, sdkerrors.ErrMarketClosed):
        return fmt.Errorf("trading failed: market is no longer active")
    default:
        return fmt.Errorf("trading failed: %v", err)
    }
}
```

## Retry Strategies

### With Exponential Backoff

```go
func withRetry(fn func() error, maxRetries int) error {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        
        if errors.Is(lastErr, sdkerrors.ErrRateLimitExceeded) {
            time.Sleep(time.Duration(1<<i) * time.Second)
            continue
        }
        
        if errors.Is(lastErr, sdkerrors.ErrCircuitOpen) {
            time.Sleep(time.Duration(1<<i) * 5 * time.Second)
            continue
        }
        
        return lastErr
    }
    return lastErr
}
```

### With Circuit Breaker

The SDK includes built-in circuit breaker support in `pkg/transport`:

```go
cfg := transport.DefaultConfig()
cfg.CircuitBreakerEnabled = true
cfg.CircuitBreakerThreshold = 5

client := polymarket.NewClient(
    polymarket.WithTransportConfig(cfg),
)
```

## Best Practices

1. **Always check errors** - Never ignore errors, even in simple use cases
2. **Use error codes** - Check `errors.Is()` with SDK error codes for programmatic handling
3. **Implement retries** - Use exponential backoff for transient failures (rate limits, network issues)
4. **Log appropriately** - Include error codes in logs for debugging
5. **Fail gracefully** - Provide user-friendly messages when errors occur
6. **Handle timeouts** - Set appropriate timeouts for network operations

```go
// Example: Production-ready error handling
func placeOrder(ctx context.Context, client clob.Client, order *clobtypes.Order) (*clobtypes.OrderResponse, error) {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        resp, err := client.CreateOrder(ctx, order)
        if err == nil {
            return resp, nil
        }
        
        switch {
        case errors.Is(err, sdkerrors.ErrRateLimitExceeded):
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        case errors.Is(err, sdkerrors.ErrInsufficientFunds),
             errors.Is(err, sdkerrors.ErrMarketClosed),
             errors.Is(err, sdkerrors.ErrGeoblocked):
            return nil, err
        default:
            return nil, err
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```
