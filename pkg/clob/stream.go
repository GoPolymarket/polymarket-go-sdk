package clob

import "github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/clob/clobtypes"

import "context"

// StreamResult wraps a streamed item or an error.
type StreamResult[T any] struct {
	Item T
	Err  error
}

// StreamFetch fetches a page of data given a cursor.
// It should return the items, next cursor, and any error.
type StreamFetch[T any] func(ctx context.Context, cursor string) ([]T, string, error)

// StreamData streams items starting from the initial cursor.
func StreamData[T any](ctx context.Context, fetch StreamFetch[T]) <-chan StreamResult[T] {
	return StreamDataWithCursor(ctx, clobtypes.InitialCursor, fetch)
}

// StreamDataWithCursor streams items starting from a specific cursor.
func StreamDataWithCursor[T any](ctx context.Context, cursor string, fetch StreamFetch[T]) <-chan StreamResult[T] {
	out := make(chan StreamResult[T], 1) // Buffered to prevent goroutine leak if consumer stops receiving
	go func() {
		defer close(out)
		if ctx == nil {
			ctx = context.Background()
		}
		if cursor == "" {
			cursor = clobtypes.InitialCursor
		}

		for cursor != clobtypes.EndCursor {
			// Check context before each fetch operation
			if err := ctx.Err(); err != nil {
				select {
				case out <- StreamResult[T]{Err: err}:
				case <-ctx.Done():
				}
				return
			}

			// Make fetch operation cancellable by passing context
			items, next, err := fetch(ctx, cursor)
			if err != nil {
				select {
				case out <- StreamResult[T]{Err: err}:
				case <-ctx.Done():
				}
				return
			}

			for _, item := range items {
				// Check context before sending each item
				if err := ctx.Err(); err != nil {
					select {
					case out <- StreamResult[T]{Err: err}:
					case <-ctx.Done():
					}
					return
				}

				// Use select to make send cancellable
				select {
				case out <- StreamResult[T]{Item: item}:
				case <-ctx.Done():
					return
				}
			}

			if next == "" || next == cursor {
				return
			}
			cursor = next
		}
	}()
	return out
}
