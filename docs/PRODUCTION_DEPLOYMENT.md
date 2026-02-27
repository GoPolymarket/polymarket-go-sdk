# Production Deployment Checklist

This checklist covers essential items for deploying applications built with the Polymarket Go SDK to production.

## Security

### Credentials Management

- [ ] Private keys stored in secure vault (AWS Secrets Manager, HashiCorp Vault)
- [ ] API keys and secrets never committed to version control
- [ ] Environment variables used for sensitive configuration
- [ ] Builder attribution keys properly secured
- [ ] Rotation policy in place for API credentials

### Network Security

- [ ] TLS 1.2+ enforced for all connections
- [ ] Rate limiting configured appropriately
- [ ] IP allowlisting for admin endpoints (if applicable)
- [ ] VPN or private network for internal services

### Code Security

- [ ] No hardcoded credentials in source code
- [ ] Dependencies audited for vulnerabilities (`go vet`, `golangci-lint`)
- [ ] Input validation on all user-facing endpoints
- [ ] Proper error handling (no stack traces exposed)

## Reliability

### Error Handling

- [ ] All errors logged with appropriate levels
- [ ] Retry logic implemented for transient failures
- [ ] Circuit breaker configured for external API calls
- [ ] Timeout set on all network requests: 30s (recommended)

### Monitoring

- [ ] Health check endpoint implemented (`/health`)
- [ ] Metrics exported (Prometheus, DataDog, etc.)
- [ ] Structured logging in place
- [ ] Alerting configured for:
  - High error rates
  - API rate limit rejections
  - Latency spikes
  - Unusual trading patterns

### Backup & Recovery

- [ ] State persistence configured (Postgres/Redis)
- [ ] Regular backups scheduled
- [ ] Recovery procedure documented
- [ ] Failover capability for high availability

## Performance

### Optimization

- [ ] Benchmark tests pass (no regression)
- [ ] Connection pooling configured
- [ ] WebSocket heartbeat enabled
- [ ] Reconnection logic tested

### Capacity Planning

- [ ] Load testing performed
- [ ] Max connections tuned for expected load
- [ ] Rate limits appropriate for use case

## Configuration

### Environment Variables

Required:
```bash
POLYMARKET_PK           # Private key (from secure vault)
POLYMARKET_API_KEY      # L2 API key
POLYMARKET_API_SECRET   # L2 API secret
POLYMARKET_API_PASSPHRASE  # L2 API passphrase
```

Optional:
```bash
POLYMARKET_BUILDER_KEY        # Builder API key (for attribution)
POLYMARKET_BUILDER_SECRET     # Builder secret
POLYMARKET_BUILDER_PASSPHRASE # Builder passphrase

# Network
HTTP_TIMEOUT=30s
WS_RECONNECT_MAX=10

# Risk Limits
MAX_ORDER_VALUE=1000
MAX_DAILY_LOSS=100
```

### Feature Flags

- [ ] Dry-run mode available for testing
- [ ] Execute flag required for live trading
- [ ] Circuit breaker can be toggled

## Operational

### Deployment

- [ ] Blue-green or canary deployment configured
- [ ] Rollback procedure tested
- [ ] Health checks pass before traffic
- [ ] Graceful shutdown implemented

### Runbook

- [ ] Deployment procedure documented
- [ ] Troubleshooting guide for common errors
- [ ] Contact information for on-call
- [ ] Escalation procedure defined

## Testing

### Pre-Production

- [ ] Unit tests pass (>40% coverage)
- [ ] Integration tests pass
- [ ] End-to-end tests pass
- [ ] Performance tests under expected load

### Validation

- [ ] Testnet trading verified
- [ ] Builder attribution confirmed
- [ ] WebSocket reconnection tested
- [ ] Error handling verified under failure scenarios

## Compliance

- [ ] Audit logging enabled
- [ ] Data retention policy in place
- [ ] Access controls documented
- [ ] Regulatory requirements met (if applicable)

## Deployment Commands

```bash
# Build
go build -o bin/server ./cmd/server

# Testnet verification
./bin/server --dry-run=true

# Production deployment
./bin/server --execute=true

# With custom config
./bin/server --config=/etc/polymarket/config.yaml
```

## Quick Reference

| Check | Priority | Impact |
|-------|----------|--------|
| Private key security | Critical | Financial loss |
| TLS enabled | Critical | Data exposure |
| Rate limiting | High | Service availability |
| Monitoring/Alerting | High | Incident detection |
| Backup strategy | High | Data recovery |
| Health checks | Medium | Deployment safety |
| Error handling | Medium | User experience |
