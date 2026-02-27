# Contributing to Polymarket Go SDK

Thank you for your interest in contributing! This guide will help you get started.

## Quick Start

```bash
# Fork and clone the repository
git clone https://github.com/GoPolymarket/polymarket-go-sdk.git
cd polymarket-go-sdk

# Install dependencies
go mod download

# Run tests
go test ./...
```

## Development Setup

### Prerequisites

- Go 1.21+
- Docker (for integration tests)
- Make

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run benchmarks
go test -bench=. ./pkg/clob/...
```

### Linting

```bash
# Install golangci-lint
make lint-install

# Run linter
make lint
```

## Code Standards

### Code Style

- Follow Go standard conventions (run `gofmt` before committing)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Testing Requirements

- All new code must include tests
- Maintain >=40% test coverage (enforced by CI)
- Use table-driven tests where appropriate
- Include benchmark tests for performance-critical code

### Commit Messages

Follow conventional commits:

```
feat: add new order type support
fix: resolve WebSocket reconnection issue
docs: update API documentation
test: add benchmark tests for order builder
refactor: simplify error handling logic
```

## Pull Request Process

### Before Submitting

1. **Run tests**: `go test ./...`
2. **Run linter**: `golangci-lint run ./...`
3. **Check coverage**: `make coverage`
4. **Format code**: `gofmt -w .`

### PR Description

Include:
- Summary of changes
- Related issue numbers
- Testing performed
- Screenshots (if UI changes)

### Review Criteria

- Code compiles without warnings
- Tests pass
- Coverage maintained or improved
- No linting errors
- Documentation updated (if needed)

## Project Structure

```
polymarket-go-sdk/
├── cmd/                    # CLI tools
├── docs/                   # Documentation
├── examples/               # Usage examples
├── pkg/
│   ├── auth/              # Authentication (signing, keys)
│   ├── bot/               # Trading bot utilities
│   ├── bridge/            # Cross-chain bridging
│   ├── clob/              # CLOB API client
│   │   └── ws/            # WebSocket client
│   ├── ctf/               # Conditional Token Framework
│   ├── data/              # Data API client
│   ├── errors/            # Error definitions
│   ├── gamma/             # Gamma API client
│   ├── rtds/              # Real-time data streaming
│   ├── transport/         # HTTP transport layer
│   └── types/             # Shared types
```

## Adding New Features

### 1. Design Phase

- Open an issue to discuss the feature
- Consider backward compatibility
- Think about error handling

### 2. Implementation

- Add tests in `*_test.go` files
- Add benchmarks in `*_bench_test.go` if applicable
- Update documentation in `docs/`
- Add examples in `examples/`

### 3. API Design Guidelines

- Use interfaces for flexibility
- Provide sensible defaults
- Include context support for cancellation
- Return structured errors

## Reporting Issues

When reporting bugs, include:
- Go version
- OS/Architecture
- Minimal reproduction code
- Expected vs actual behavior

## Getting Help

- Open an issue for bugs/features
- Join discussions in the community
- Check existing examples and documentation

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
