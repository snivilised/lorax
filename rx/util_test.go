package rx_test

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

func channelValueN[T any](ctx context.Context, items ...any) chan rx.Item[T] {
	next := make(chan rx.Item[T])
	go func() {
		for _, item := range items {
			switch item := item.(type) {
			case rx.Item[T]:
				item.SendContext(ctx, next)

			case error:
				rx.Error[T](item).SendContext(ctx, next)

			case int:
				rx.Num[T](item).SendContext(ctx, next)

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

func testObservableN[T any](ctx context.Context, items ...any) rx.Observable[T] {
	// items is a collection of any because we need the ability to send a stream
	// of events that may include errors; 1, 2, err, 4, ..., without enforcing
	// that the client should manufacture Item[T]s; Of(1), Of(2), Error(err), Of(4).
	//
	return rx.FromChannel(channelValueN[T](ctx, convertAllItemsToAny(items)...))
}
