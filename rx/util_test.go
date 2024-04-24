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

func testObservableWith[T any](ctx context.Context, items ...any) func(opts ...rx.Option[T]) rx.Observable[T] {
	// items is a collection of any because we need the ability to send a stream
	// of events that may include errors; 1, 2, err, 4, ..., without enforcing
	// that the client should manufacture Item[T]s; Of(1), Of(2), Error(err), Of(4).
	//
	return func(opts ...rx.Option[T]) rx.Observable[T] {
		return rx.FromChannel(channelValue[T](ctx, convertAllItemsToAny(items)...), opts...)
	}
}

type widget struct {
	id     int
	name   string
	amount int
}

func (w widget) Field() int {
	return w.id
}

func (w widget) Inc(index *widget, by widget) *widget {
	// we can't increment w.wid, because Inc is a non pointer receiver,
	// therefore, as a work-around, we must increment index which is a
	// back-door copy of the original instance.
	//
	index.id += by.id

	return index
}

func (w widget) Index(i int) *widget {
	return &widget{
		id: i,
	}
}

type widgetByIDRangeIterator struct {
	StartAt widget
	By      widget
	Whilst  func(current widget) bool
	zero    widget
}

func (i *widgetByIDRangeIterator) Init() error {
	return nil
}

// Start should return the initial index value. If the By value has
// not been set, it will default to 1.
func (i *widgetByIDRangeIterator) Start() (*widget, error) {
	if i.By.id == 0 {
		i.By = widget{id: 1}
	}

	if i.Whilst == nil {
		return &i.zero, rx.BadRangeIteratorError{}
	}

	index := i.StartAt

	return &index, nil
}

func (i *widgetByIDRangeIterator) Step() int {
	return i.By.id
}

// Increment returns a pointer to a new instance of with incremented index value
func (i *widgetByIDRangeIterator) Increment(index *widget) *widget {
	index.id += i.By.id

	return index
}

// While defines a condition that must be true for the loop to
// continue iterating.
func (i *widgetByIDRangeIterator) While(current widget) bool {
	return i.Whilst(current)
}

func widgetLessThan(until widget) func(widget) bool {
	return func(current widget) bool {
		return current.id < until.id
	}
}
