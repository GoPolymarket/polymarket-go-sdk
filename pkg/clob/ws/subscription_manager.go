package ws

import (
	"sync"
	"sync/atomic"
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
	closed    atomic.Bool
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
	if s.closed.Load() {
		return
	}
	defer func() {
		_ = recover()
	}()
	select {
	case s.ch <- msg:
		return
	default:
		s.notifyLag(1)
	}
}

func (s *subscriptionEntry[T]) notifyLag(count int) {
	if count <= 0 {
		return
	}
	select {
	case s.errCh <- LaggedError{Count: count, Channel: s.channel, EventType: s.event}:
	default:
	}
}

func (s *subscriptionEntry[T]) close() bool {
	if s.closed.Swap(true) {
		return false
	}
	s.closeOnce.Do(func() {
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
