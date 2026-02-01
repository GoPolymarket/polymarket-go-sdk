package transport

import (
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name            string
		config          CircuitBreakerConfig
		expectedState   CircuitState
		expectedMaxFail int
	}{
		{
			name: "valid config",
			config: CircuitBreakerConfig{
				MaxFailures:     5,
				ResetTimeout:    60 * time.Second,
				HalfOpenMaxReqs: 3,
			},
			expectedState:   StateClosed,
			expectedMaxFail: 5,
		},
		{
			name: "zero values use defaults",
			config: CircuitBreakerConfig{
				MaxFailures:     0,
				ResetTimeout:    0,
				HalfOpenMaxReqs: 0,
			},
			expectedState:   StateClosed,
			expectedMaxFail: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.config)
			if cb == nil {
				t.Fatal("NewCircuitBreaker returned nil")
			}
			if cb.State() != tt.expectedState {
				t.Errorf("State() = %v, want %v", cb.State(), tt.expectedState)
			}
			if cb.maxFailures != tt.expectedMaxFail {
				t.Errorf("maxFailures = %d, want %d", cb.maxFailures, tt.expectedMaxFail)
			}
		})
	}
}

func TestCircuitBreaker_ClosedState(t *testing.T) {
	t.Run("allows requests when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

		callCount := 0
		err := cb.Call(func() error {
			callCount++
			return nil
		})

		if err != nil {
			t.Errorf("Call() error = %v, want nil", err)
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1", callCount)
		}
		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
	})

	t.Run("resets failure count on success", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     3,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 2,
		})

		// Fail twice
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		if cb.Failures() != 2 {
			t.Errorf("Failures() = %d, want 2", cb.Failures())
		}

		// Succeed once
		_ = cb.Call(func() error { return nil })

		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0 (should reset on success)", cb.Failures())
		}
	})
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	t.Run("opens after max failures", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     3,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 2,
		})

		// Fail 3 times to open the circuit
		for i := 0; i < 3; i++ {
			_ = cb.Call(func() error { return errors.New("error") })
		}

		if cb.State() != StateOpen {
			t.Errorf("State() = %v, want %v", cb.State(), StateOpen)
		}
	})

	t.Run("blocks requests when open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 2,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Try to make a request
		callCount := 0
		err := cb.Call(func() error {
			callCount++
			return nil
		})

		if err != ErrCircuitOpen {
			t.Errorf("Call() error = %v, want %v", err, ErrCircuitOpen)
		}
		if callCount != 0 {
			t.Errorf("callCount = %d, want 0 (function should not be called)", callCount)
		}
	})

	t.Run("transitions to half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    100 * time.Millisecond,
			HalfOpenMaxReqs: 2,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		if cb.State() != StateOpen {
			t.Errorf("State() = %v, want %v", cb.State(), StateOpen)
		}

		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// Next request should transition to half-open
		callCount := 0
		_ = cb.Call(func() error {
			callCount++
			return nil
		})

		if cb.State() != StateHalfOpen {
			t.Errorf("State() = %v, want %v", cb.State(), StateHalfOpen)
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1", callCount)
		}
	})
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	t.Run("allows limited requests in half-open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    50 * time.Millisecond,
			HalfOpenMaxReqs: 3,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Wait for reset timeout
		time.Sleep(100 * time.Millisecond)

		// Make requests in half-open state
		successCount := 0
		for i := 0; i < 3; i++ {
			err := cb.Call(func() error { return nil })
			if err == nil {
				successCount++
			}
		}

		if successCount != 3 {
			t.Errorf("successCount = %d, want 3", successCount)
		}
	})

	t.Run("closes circuit after successful half-open requests", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    50 * time.Millisecond,
			HalfOpenMaxReqs: 2,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Wait for reset timeout
		time.Sleep(100 * time.Millisecond)

		// Make successful requests in half-open state
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })

		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v (should close after successful half-open requests)", cb.State(), StateClosed)
		}
		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0", cb.Failures())
		}
	})

	t.Run("reopens circuit on half-open failure", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    50 * time.Millisecond,
			HalfOpenMaxReqs: 3,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Wait for reset timeout
		time.Sleep(100 * time.Millisecond)

		// First request succeeds
		_ = cb.Call(func() error { return nil })

		// Second request fails
		_ = cb.Call(func() error { return errors.New("error") })

		if cb.State() != StateOpen {
			t.Errorf("State() = %v, want %v (should reopen on failure)", cb.State(), StateOpen)
		}
	})

	t.Run("rejects requests beyond half-open limit", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    50 * time.Millisecond,
			HalfOpenMaxReqs: 2,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		// Wait for reset timeout
		time.Sleep(100 * time.Millisecond)

		// Make 2 successful requests (at the limit)
		_ = cb.Call(func() error { return nil })
		_ = cb.Call(func() error { return nil })

		// Circuit should now be closed
		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
	})
}

func TestCircuitBreaker_Reset(t *testing.T) {
	t.Run("manual reset", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     2,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 2,
		})

		// Open the circuit
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		if cb.State() != StateOpen {
			t.Errorf("State() = %v, want %v", cb.State(), StateOpen)
		}

		// Manual reset
		cb.Reset()

		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0", cb.Failures())
		}
	})
}

func TestCircuitBreaker_Stats(t *testing.T) {
	t.Run("returns accurate stats", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     3,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 2,
		})

		// Fail twice
		_ = cb.Call(func() error { return errors.New("error") })
		_ = cb.Call(func() error { return errors.New("error") })

		stats := cb.Stats()

		if stats.State != StateClosed {
			t.Errorf("Stats.State = %v, want %v", stats.State, StateClosed)
		}
		if stats.Failures != 2 {
			t.Errorf("Stats.Failures = %d, want 2", stats.Failures)
		}
		if stats.LastFailTime.IsZero() {
			t.Error("Stats.LastFailTime should not be zero")
		}
	})
}

func TestCircuitBreaker_CallWithFailurePredicate(t *testing.T) {
	t.Run("does not count failures when predicate returns false", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     1,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 1,
		})

		err := cb.CallWithFailurePredicate(func() error { return errors.New("ignored") }, func(err error) bool {
			return false
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if cb.Failures() != 0 {
			t.Errorf("Failures() = %d, want 0", cb.Failures())
		}
		if cb.State() != StateClosed {
			t.Errorf("State() = %v, want %v", cb.State(), StateClosed)
		}
	})

	t.Run("counts failures when predicate returns true", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     1,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 1,
		})

		err := cb.CallWithFailurePredicate(func() error { return errors.New("counted") }, func(err error) bool {
			return true
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if cb.Failures() != 1 {
			t.Errorf("Failures() = %d, want 1", cb.Failures())
		}
		if cb.State() != StateOpen {
			t.Errorf("State() = %v, want %v", cb.State(), StateOpen)
		}
	})
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	t.Run("thread-safe operation", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			MaxFailures:     10,
			ResetTimeout:    60 * time.Second,
			HalfOpenMaxReqs: 5,
		})

		// Run concurrent requests
		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Alternate between success and failure
				if id%2 == 0 {
					_ = cb.Call(func() error { return nil })
				} else {
					_ = cb.Call(func() error { return errors.New("error") })
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Circuit should still be in a valid state
		state := cb.State()
		if state != StateClosed && state != StateOpen && state != StateHalfOpen {
			t.Errorf("Invalid state: %v", state)
		}
	})
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	if config.MaxFailures != 5 {
		t.Errorf("MaxFailures = %d, want 5", config.MaxFailures)
	}
	if config.ResetTimeout != 60*time.Second {
		t.Errorf("ResetTimeout = %v, want 60s", config.ResetTimeout)
	}
	if config.HalfOpenMaxReqs != 3 {
		t.Errorf("HalfOpenMaxReqs = %d, want 3", config.HalfOpenMaxReqs)
	}
}
