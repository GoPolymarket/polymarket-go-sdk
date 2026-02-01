package transport

import (
	"sync"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/errors"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState string

const (
	// StateClosed means requests are allowed through.
	StateClosed CircuitState = "closed"
	// StateOpen means requests are blocked.
	StateOpen CircuitState = "open"
	// StateHalfOpen means a limited number of requests are allowed to test if the service has recovered.
	StateHalfOpen CircuitState = "half-open"
)

// Use unified error definitions from pkg/errors
var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.ErrCircuitOpen
	// ErrTooManyRequests is returned when too many requests are made in half-open state.
	ErrTooManyRequests = errors.ErrTooManyRequests
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	halfOpenMaxReqs int

	mu              sync.RWMutex
	state           CircuitState
	failures        int
	lastFailTime    time.Time
	halfOpenReqs    int
	halfOpenSuccess int
	halfOpenFailure int
}

// FailurePredicate determines whether an error should count as a circuit breaker failure.
type FailurePredicate func(error) bool

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	MaxFailures     int           // Number of failures before opening the circuit
	ResetTimeout    time.Duration // Time to wait before transitioning from open to half-open
	HalfOpenMaxReqs int           // Maximum requests allowed in half-open state
}

// DefaultCircuitBreakerConfig returns a default configuration.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:     5,
		ResetTimeout:    60 * time.Second,
		HalfOpenMaxReqs: 3,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 60 * time.Second
	}
	if config.HalfOpenMaxReqs <= 0 {
		config.HalfOpenMaxReqs = 3
	}

	return &CircuitBreaker{
		maxFailures:     config.MaxFailures,
		resetTimeout:    config.ResetTimeout,
		halfOpenMaxReqs: config.HalfOpenMaxReqs,
		state:           StateClosed,
	}
}

// Call executes the given function if the circuit breaker allows it.
func (cb *CircuitBreaker) Call(fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	cb.afterRequest(err)
	return err
}

// CallWithFailurePredicate executes the given function and records failures based on the predicate.
// If shouldCount returns false, the error is treated as a success for breaker statistics.
func (cb *CircuitBreaker) CallWithFailurePredicate(fn func() error, shouldCount FailurePredicate) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()
	countFailure := err != nil
	if countFailure && shouldCount != nil {
		countFailure = shouldCount(err)
	}
	cb.afterRequestWithPolicy(err, countFailure)
	return err
}

// beforeRequest checks if the request should be allowed.
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) <= cb.resetTimeout {
			return ErrCircuitOpen
		}
		cb.state = StateHalfOpen
		cb.halfOpenReqs = 0
		cb.halfOpenSuccess = 0
		cb.halfOpenFailure = 0
		fallthrough

	case StateHalfOpen:
		if cb.halfOpenReqs >= cb.halfOpenMaxReqs {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil

	default:
		return nil
	}
}

// afterRequest records the result of the request.
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.afterRequestWithPolicy(err, true)
}

func (cb *CircuitBreaker) afterRequestWithPolicy(err error, countFailure bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		if countFailure {
			cb.recordFailure()
		} else {
			cb.recordSuccess()
		}
		return
	}
	cb.recordSuccess()
}

// recordFailure records a failed request.
func (cb *CircuitBreaker) recordFailure() {
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

	case StateHalfOpen:
		cb.halfOpenFailure++
		// If any request fails in half-open, go back to open
		cb.state = StateOpen
		cb.failures = cb.maxFailures
	}
}

// recordSuccess records a successful request.
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.halfOpenSuccess++
		// If all half-open requests succeed, close the circuit
		if cb.halfOpenSuccess >= cb.halfOpenMaxReqs {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenReqs = 0
	cb.halfOpenSuccess = 0
	cb.halfOpenFailure = 0
}

// Stats returns statistics about the circuit breaker.
type CircuitBreakerStats struct {
	State           CircuitState
	Failures        int
	HalfOpenReqs    int
	HalfOpenSuccess int
	HalfOpenFailure int
	LastFailTime    time.Time
}

// Stats returns current statistics.
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return CircuitBreakerStats{
		State:           cb.state,
		Failures:        cb.failures,
		HalfOpenReqs:    cb.halfOpenReqs,
		HalfOpenSuccess: cb.halfOpenSuccess,
		HalfOpenFailure: cb.halfOpenFailure,
		LastFailTime:    cb.lastFailTime,
	}
}
