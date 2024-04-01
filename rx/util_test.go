package rx_test

import (
	"context"
	"errors"

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
			}
		}

		close(next)
	}()

	return next
}

func convertAllItemsToAny[T any](items []T) []any {
	return lo.Map(items, func(it T, _ int) any {
		return it
	})
}

func testObservable[T any](ctx context.Context, items ...any) rx.Observable[T] {
	return rx.FromChannel(channelValue[T](ctx, convertAllItemsToAny(items)...))
}
