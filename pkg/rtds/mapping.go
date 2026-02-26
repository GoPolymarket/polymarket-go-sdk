package rtds

import (
	"bytes"
	"encoding/json"
	"strings"
)

func symbolSet(symbols []string) map[string]struct{} {
	if len(symbols) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(symbols))
	for _, s := range symbols {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			continue
		}
		set[s] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func mapStream[T any](src *Stream[RtdsMessage], topic, msgType string, mapFn func(RtdsMessage) (T, bool)) *Stream[T] {
	outC := make(chan T, defaultStreamBuffer)
	errC := make(chan error, defaultErrBuffer)

	go func() {
		defer close(outC)
		defer close(errC)
		for {
			select {
			case msg, ok := <-src.C:
				if !ok {
					return
				}
				mapped, ok := mapFn(msg)
				if !ok {
					continue
				}
				select {
				case outC <- mapped:
				default:
					select {
					case errC <- LaggedError{Count: 1, Topic: topic, MsgType: msgType}:
					default:
					}
				}
			case err, ok := <-src.Err:
				if !ok {
					return
				}
				select {
				case errC <- err:
				default:
				}
			}
		}
	}()

	return &Stream[T]{
		C:      outC,
		Err:    errC,
		closeF: src.Close,
	}
}

func parseMessages(message []byte) ([]RtdsMessage, error) {
	trimmed := bytes.TrimSpace(message)
	if len(trimmed) == 0 {
		return nil, nil
	}
	if trimmed[0] == '[' {
		var msgs []RtdsMessage
		if err := json.Unmarshal(trimmed, &msgs); err != nil {
			return nil, err
		}
		return msgs, nil
	}
	var msg RtdsMessage
	if err := json.Unmarshal(trimmed, &msg); err != nil {
		return nil, err
	}
	return []RtdsMessage{msg}, nil
}
