package ws

import (
	"sync"
)

const (
	defaultStreamBuffer = 100
	defaultErrBuffer    = 10
)

type subscriptionEntry[T any] struct {
	id        string
	channel   Channel
	event     EventType
	assets    map[string]struct{}
	markets   map[string]struct{}
	ch        chan T
	errCh     chan error
	mu        sync.RWMutex // Protects channel operations
	closed    bool
	closeOnce sync.Once
}

func (s *subscriptionEntry[T]) matchesAsset(assetID string) bool {
	if len(s.assets) == 0 {
		return true
	}
	_, ok := s.assets[assetID]
	return ok
}

func (s *subscriptionEntry[T]) matchesAnyAsset(assetIDs []string) bool {
	if len(s.assets) == 0 {
		return true
	}
	for _, id := range assetIDs {
		if _, ok := s.assets[id]; ok {
			return true
		}
	}
	return false
}

func (s *subscriptionEntry[T]) matchesMarket(market string) bool {
	if len(s.markets) == 0 {
		return true
	}
	_, ok := s.markets[market]
	return ok
}

func (s *subscriptionEntry[T]) trySend(msg T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}
	// Use non-blocking send to prevent blocking
	select {
	case s.ch <- msg:
		return
	default:
		s.notifyLagLocked(1)
	}
}

// notifyLagLocked sends a lag notification without acquiring the lock.
// Caller must hold at least s.mu.RLock().
func (s *subscriptionEntry[T]) notifyLagLocked(count int) {
	if count <= 0 || s.closed {
		return
	}
	select {
	case s.errCh <- LaggedError{Count: count, Channel: s.channel, EventType: s.event}:
	default:
	}
}

func (s *subscriptionEntry[T]) notifyLag(count int) {
	if count <= 0 {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}
	select {
	case s.errCh <- LaggedError{Count: count, Channel: s.channel, EventType: s.event}:
	default:
	}
}

func (s *subscriptionEntry[T]) close() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return false
	}

	s.closeOnce.Do(func() {
		s.closed = true
		close(s.ch)
		close(s.errCh)
	})
	return true
}

func makeIDSet(ids []string) map[string]struct{} {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		set[id] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func snapshotSubs[T any](subs map[string]*subscriptionEntry[T]) []*subscriptionEntry[T] {
	out := make([]*subscriptionEntry[T], 0, len(subs))
	for _, sub := range subs {
		out = append(out, sub)
	}
	return out
}

func closeSubMap[T any](subs map[string]*subscriptionEntry[T]) {
	for _, sub := range subs {
		sub.close()
	}
	for key := range subs {
		delete(subs, key)
	}
}
