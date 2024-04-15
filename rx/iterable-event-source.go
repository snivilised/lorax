package rx

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
	"sync"

	"github.com/snivilised/lorax/enums"
)

type eventSourceIterable[T any] struct {
	sync.RWMutex
	observers []chan Item[T]
	disposed  bool
	opts      []Option[T]
}

func newEventSourceIterable[T any](ctx context.Context,
	next <-chan Item[T], strategy enums.BackPressureStrategy, opts ...Option[T],
) Iterable[T] {
	it := &eventSourceIterable[T]{
		observers: make([]chan Item[T], 0),
		opts:      opts,
	}

	go func() {
		defer func() {
			it.closeAllObservers()
		}()

		deliver := func(item Item[T]) (done bool) {
			it.RLock()
			defer it.RUnlock()

			switch strategy {
			default:
				fallthrough
			case enums.Block:
				for _, observer := range it.observers {
					if !item.SendContext(ctx, observer) {
						return true
					}
				}
			case enums.Drop:
				for _, observer := range it.observers {
					select {
					default:
					case <-ctx.Done():
						return true
					case observer <- item:
					}
				}
			}

			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-next:
				if !ok {
					return
				}

				if done := deliver(item); done {
					return
				}
			}
		}
	}()

	return it
}

func (i *eventSourceIterable[T]) closeAllObservers() {
	i.Lock()
	for _, observer := range i.observers {
		close(observer)
	}

	i.disposed = true
	i.Unlock()
}

func (i *eventSourceIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()

	i.Lock()
	if i.disposed {
		close(next)
	} else {
		i.observers = append(i.observers, next)
	}
	i.Unlock()

	return next
}
