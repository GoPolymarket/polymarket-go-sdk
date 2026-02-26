package rtds

import (
	"sync"
	"sync/atomic"
)

type subscriptionEntry struct {
	id        string
	key       string
	topic     string
	msgType   string
	filter    func(RtdsMessage) bool
	ch        chan RtdsMessage
	errCh     chan error
	closed    atomic.Bool
	closeOnce sync.Once
}

func (s *subscriptionEntry) matches(msg RtdsMessage) bool {
	if msg.Topic != s.topic {
		return false
	}
	if s.msgType != "*" && msg.MsgType != s.msgType {
		return false
	}
	if s.filter != nil {
		return s.filter(msg)
	}
	return true
}

func (s *subscriptionEntry) trySend(msg RtdsMessage) {
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

func (s *subscriptionEntry) notifyLag(count int) {
	if count <= 0 {
		return
	}
	err := LaggedError{Count: count, Topic: s.topic, MsgType: s.msgType}
	select {
	case s.errCh <- err:
	default:
	}
}

func (s *subscriptionEntry) close() {
	if s.closed.Swap(true) {
		return
	}
	s.closeOnce.Do(func() {
		close(s.ch)
		close(s.errCh)
	})
}
