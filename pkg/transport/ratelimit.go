package transport

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu             sync.Mutex
	capacity       int
	tokensPerSec   float64
	tokens         float64
	lastRefill     time.Time
	stopped        bool
	stopOnce       sync.Once
}

// NewRateLimiter creates a new rate limiter with the specified requests per second.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // Default to 10 requests per second
	}

	return &RateLimiter{
		capacity:     requestsPerSecond,
		tokensPerSec: float64(requestsPerSecond),
		tokens:       float64(requestsPerSecond),
		lastRefill:   time.Now(),
	}
}

// Start begins the token refill process (no-op for timestamp-based implementation).
func (rl *RateLimiter) Start() {
	// No-op: timestamp-based rate limiter doesn't need a background goroutine
}

// Wait blocks until a token is available or the context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()

	// Refill tokens based on elapsed time
	rl.refillTokens()

	// If we have tokens available, consume one and return immediately
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		rl.mu.Unlock()
		return nil
	}

	// Calculate how long to wait for the next token
	tokensNeeded := 1.0 - rl.tokens
	waitDuration := time.Duration(float64(time.Second) * tokensNeeded / rl.tokensPerSec)

	// Unlock while waiting to allow other goroutines to proceed
	rl.mu.Unlock()

	// Wait for either the required duration or context cancellation
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Re-acquire lock and consume token
		rl.mu.Lock()
		rl.refillTokens()
		rl.tokens -= 1.0
		rl.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (rl *RateLimiter) TryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}
	return false
}

// Stop stops the token refill process.
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.stopped = true
}

// refillTokens calculates and adds tokens based on elapsed time since last refill.
// Must be called with rl.mu held.
func (rl *RateLimiter) refillTokens() {
	// Don't refill if stopped
	if rl.stopped {
		return
	}

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Calculate tokens to add based on elapsed time
	tokensToAdd := elapsed.Seconds() * rl.tokensPerSec

	// Add tokens up to capacity
	rl.tokens += tokensToAdd
	if rl.tokens > float64(rl.capacity) {
		rl.tokens = float64(rl.capacity)
	}

	rl.lastRefill = now
}

// Capacity returns the maximum number of tokens the bucket can hold.
func (rl *RateLimiter) Capacity() int {
	return rl.capacity
}

// Available returns the approximate number of tokens currently available.
// This is an estimate and may not be exact due to concurrent access.
func (rl *RateLimiter) Available() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()
	return int(rl.tokens)
}
