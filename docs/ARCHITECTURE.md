# Architecture & Design

This SDK follows a modular, "institutional-first" architecture designed for high maintainability and alignment with Polymarket's official Rust client (`rs-clob-client`).

## High-Level Overview

The codebase is organized by domain domain rather than by layer. This ensures that features like RFQ or WebSocket streaming can be versioned and maintained independently.

```text
pkg/
├── auth/              # Authentication & Signing
│   ├── kms/           # AWS KMS Integration (EIP-712)
│   └── ...
├── clob/              # Core Trading Logic
│   ├── client.go      # REST Interface
│   ├── clobtypes/     # Shared Data Models (Order, Market, etc.)
│   ├── rfq/           # Institutional RFQ Module
│   ├── ws/            # WebSocket Subsystem
│   └── heartbeat/     # Liveness Manager
└── ...
```

## Key Design Decisions

### 1. Subsystem Delegation
Instead of a monolithic `Client` struct with hundreds of methods, we use a delegation pattern.
- **REST**: `client.CLOB()`
- **RFQ**: `client.CLOB().RFQ()`
- **WebSocket**: `client.CLOB().WS()`

This keeps the API surface area clean and discoverable (IntelliSense friendly).

### 2. Type Decoupling (`clobtypes`)
We introduced `pkg/clob/clobtypes` to prevent circular dependencies between the `clob` package and its sub-modules (`rfq`, `ws`). This allows the RFQ module to reuse core Order types without importing the heavy REST client implementation.

### 3. Remote Builder Attribution
To support the ecosystem sustainably, we implemented a **Secure Remote Signing** architecture.
- **Problem**: Builder API Secrets cannot be open-sourced.
- **Solution**: The SDK sends request metadata to a verified remote signer (hosted on Zeabur).
- **Result**: Users can opt-in to support the SDK maintenance without exposing any credentials or risking their own funds.

### 4. Enterprise Security
We treat security as a first-class citizen.
- **AWS KMS**: Implemented native support for AWS KMS signing, including the complex ASN.1 to Ethereum signature conversion logic (R/S/V recovery).
- **Non-Custodial**: Private keys never need to touch the application memory if using KMS.
