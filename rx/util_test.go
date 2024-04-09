package rx_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/snivilised/lorax/rx"
)

var (
	errFoo = errors.New("foo")
	errBar = errors.New("bar")
)

func channelValue[T any](ctx context.Context, items ...any) chan rx.Item[T] {
	next := make(chan rx.Item[T])
	go func() {
		for _, item := range items {
			switch item := item.(type) {
			case rx.Item[T]:
				item.SendContext(ctx, next)

			case error:
				rx.Error[T](item).SendContext(ctx, next)

			case T:
				rx.Of(item).SendContext(ctx, next)

			default:
				// This error can occur, if the client instantiates with type T,
				// but sends a value not of type T, eg, instantiate with float32,
				// but send 42 through the channel, instead of float32(42).
				//
				err := fmt.Errorf("channel value: '%v' not sent (wrong type?)", item)

				rx.Error[T](err).SendContext(ctx, next)
			}
		}

		close(next)
	}()

	return next
}

func convertAllItemsToAny[T any](values []T) []any {
	return lo.Map(values, func(value T, _ int) any {
		return value
	})
}

func testObservable[T any](ctx context.Context, items ...any) rx.Observable[T] {
	// items is a collection of any because we need the ability to send a stream
	// of events that may include errors; 1, 2, err, 4, ..., without enforcing
	// that the client should manufacture Item[T]s; Of(1), Of(2), Error(err), Of(4).
	//
	return rx.FromChannel(channelValue[T](ctx, convertAllItemsToAny(items)...))
}

type widget struct {
	name   string
	amount int
}
