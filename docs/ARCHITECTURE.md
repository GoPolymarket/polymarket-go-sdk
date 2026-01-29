# Architecture & Design

This SDK follows a modular, "institutional-first" architecture designed for high maintainability and alignment with Polymarket's official Rust client (`rs-clob-client`).

## High-Level Overview

The codebase is organized by domain rather than by layer. This ensures that features like RFQ or WebSocket streaming can be versioned and maintained independently.

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

## Architecture Principles

1. **Separation by domain**: Each trading capability (CLOB, RFQ, WS, RTDS) is a first-class module to keep iteration speed high.
2. **Unified auth and transport**: A single client instance holds transport, auth, and config, minimizing duplicated code paths.
3. **Extensibility**: New networks or data providers can be added under `pkg/` without refactoring core APIs.
4. **Operational safety**: Connection management and retry logic live close to transport, making observability easier.

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

## 技术路线（Roadmap）

- **短期**：完善 CLOB/RFQ/WS 的稳定性与错误处理；补齐安全审计与 CI 安全扫描。
- **中期**：多云 KMS（GCP/Azure）集成；更细粒度的访问控制与请求签名策略。
- **长期**：可插拔策略与数据管道，支持自建撮合或数据中台集成。

## 参考使用场景

- **量化机构**：通过统一 API 管理多策略交易、历史与实时数据。
- **风控团队**：通过心跳与实时订阅监控市场波动。
- **数据平台**：通过 Gamma/RTDS 进行数据汇聚与合并。

## Quickstart Architecture Flow

1. 创建 `Client` 并初始化认证信息。
2. 通过 `CLOB()` 访问 REST 与 RFQ 交易接口。
3. 通过 `WS()` 建立 WebSocket 订阅。
4. 在生产环境开启 KMS 与安全审计流程（详见 [docs/SECURITY.md](SECURITY.md) 与 [docs/SECURITY_AUDIT.md](SECURITY_AUDIT.md)）。
