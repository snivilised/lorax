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
	"sync/atomic"
	"time"

	"github.com/samber/lo"
)

// Amb takes several Observables, emit all of the items from only the first of these Observables
// to emit an item or notification. (What the hell is an Amb, WTF)
func Amb[T any](observables []Observable[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()
	once := sync.Once{}

	f := func(o Observable[T]) {
		it := o.Observe(opts...)

		select {
		case <-ctx.Done():
			return
		case item, ok := <-it:
			if !ok {
				return
			}

			once.Do(func() {
				defer close(next)

				if item.IsError() {
					next <- item

					return
				}

				next <- item

				for {
					select {
					case <-ctx.Done():
						return
					case item, ok := <-it:
						if !ok {
							return
						}

						if item.IsError() {
							next <- item

							return
						}
						next <- item
					}
				}
			})
		}
	}

	for _, o := range observables {
		go f(o)
	}

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// CombineLatest combines the latest item emitted by each Observable via a specified function
// and emit items based on the results of this function. Requires a calculator, so specify
// this with the WithCalc option.
func CombineLatest[T any](f FuncN[T], observables []Observable[T],
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()
	calc := option.calc()

	go func() {
		var counter uint32

		size := uint32(len(observables))
		s := make([]T, size)
		mutex := sync.Mutex{}
		errCh := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(int(size))

		handler := func(ctx context.Context, it Iterable[T], i int) {
			defer wg.Done()

			observe := it.Observe(opts...)

			if calc == nil {
				Error[T](MissingCalcError{}).SendContext(ctx, next)

				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-observe:
					if !ok {
						return
					}

					if item.IsError() {
						next <- item
						errCh <- struct{}{}

						return
					}

					if calc.IsZero(s[i]) {
						atomic.AddUint32(&counter, 1)
					}

					mutex.Lock()
					s[i] = item.V

					if atomic.LoadUint32(&counter) == size {
						next <- Of(f(s...))
					}
					mutex.Unlock()
				}
			}
		}

		cancelCtx, cancel := context.WithCancel(ctx)

		for i, o := range observables {
			go handler(cancelCtx, o, i)
		}

		go func() {
			for range errCh {
				cancel()
			}
		}()

		wg.Wait()
		close(next)
		close(errCh)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Concat emits the emissions from two or more Observables without interleaving them.
func Concat[T any](observables []Observable[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()

	go func() {
		defer close(next)

		for _, obs := range observables {
			observe := obs.Observe(opts...)
		loop:
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-observe:
					if !ok {
						break loop
					}
					if item.IsError() {
						next <- item

						return
					}
					next <- item
				}
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Create creates an Observable from scratch by calling observer methods programmatically.
func Create[T any](f []Producer[T], opts ...Option[T]) Observable[T] {
	return &ObservableImpl[T]{
		iterable: newCreateIterable(f, opts...),
	}
}

// Defer does not create the Observable until the observer subscribes,
// and creates a fresh Observable for each observer. This creates a cold
// observable.
func Defer[T any](f []Producer[T], opts ...Option[T]) Observable[T] {
	return &ObservableImpl[T]{
		iterable: newDeferIterable(f, opts...),
	}
}

// Empty creates an Observable with no item and terminate immediately.
func Empty[T any]() Observable[T] {
	next := make(chan Item[T])
	close(next)

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// FromChannel creates a cold observable from a channel.
func FromChannel[T any](next <-chan Item[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)

	return &ObservableImpl[T]{
		parent:   ctx,
		iterable: newChannelIterable(next, opts...),
	}
}

// FromEventSource creates a hot observable from a channel.
func FromEventSource[T any](next <-chan Item[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)

	return &ObservableImpl[T]{
		iterable: newEventSourceIterable(option.buildContext(emptyContext),
			next, option.getBackPressureStrategy(),
		),
	}
}

// Interval creates an Observable emitting incremental integers infinitely between
// each given time interval.
func Interval[T any](interval Duration, opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(emptyContext)

	go func() {
		i := 0

		for {
			select {
			case <-time.After(interval.duration()):
				if !TV[T](i).SendContext(ctx, next) {
					return
				}

				i++
			case <-ctx.Done():
				close(next)
				return
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newEventSourceIterable(ctx, next, option.getBackPressureStrategy()),
	}
}

// Just creates an Observable with the provided items.
func Just[T any](values ...T) func(opts ...Option[T]) Observable[T] {
	return func(opts ...Option[T]) Observable[T] {
		return &ObservableImpl[T]{
			iterable: newJustIterable[T](lo.Map(values, func(value T, _ int) any {
				return value
			})...)(opts...),
		}
	}
}

// JustSingle is like JustItem in that it is defined for a single item iterable
// but behaves like Just in that it returns a func.
// This is probably not required, just defined for experimental purposes for now.
func JustSingle[T any](value T, opts ...Option[T]) func(opts ...Option[T]) Single[T] {
	return func(_ ...Option[T]) Single[T] {
		return &SingleImpl[T]{
			iterable: newJustIterable[T](value)(opts...),
		}
	}
}

// JustItem creates a single from one item.
func JustItem[T any](value T, opts ...Option[T]) Single[T] {
	// Why does this not return a func, but Just does?
	//
	return &SingleImpl[T]{
		iterable: newJustIterable[T](value)(opts...),
	}
}

// Just creates an Observable with the provided items.
func JustError[T any](err error) func(opts ...Option[T]) Single[T] {
	return func(opts ...Option[T]) Single[T] {
		return &SingleImpl[T]{
			iterable: newJustIterable[T](err)(opts...),
		}
	}
}

// Merge combines multiple Observables into one by merging their emissions
func Merge[T any](observables []Observable[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()
	wg := sync.WaitGroup{}
	wg.Add(len(observables))

	f := func(o Observable[T]) {
		defer wg.Done()

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-observe:
				if !ok {
					return
				}

				if item.IsError() {
					next <- item

					return
				}
				next <- item
			}
		}
	}

	for _, o := range observables {
		go f(o)
	}

	go func() {
		wg.Wait()

		close(next)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Never creates an Observable that emits no items and does not terminate.
func Never[T any]() Observable[T] {
	next := make(chan Item[T])

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Range creates an Observable that emits count sequential integers beginning
// at start.
func Range[T Numeric](iterator RangeIterator[T], opts ...Option[T]) Observable[T] {
	if err := iterator.Init(); err != nil {
		return Thrown[T](err)
	}

	return &ObservableImpl[T]{
		iterable: newRangeIterable(iterator, opts...),
	}
}

// RangePF creates an Observable that emits count sequential integers beginning
// at start, for non numeric types, which do contain a nominated proxy Numeric member
func RangePF[T ProxyField[T, O], O Numeric](iterator RangeIteratorPF[T, O],
	opts ...Option[T],
) Observable[T] {
	if err := iterator.Init(); err != nil {
		return Thrown[T](err)
	}

	return &ObservableImpl[T]{
		iterable: newRangeIterableNF(iterator, opts...),
	}
}

// Start creates an Observable from one or more directive-like Supplier
// and emits the result of each operation asynchronously on a new Observable.
func Start[T any](fs []Supplier[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(emptyContext)

	go func() {
		defer close(next)

		for _, f := range fs {
			select {
			case <-ctx.Done():
				return
			case next <- f(ctx):
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Thrown creates an Observable that emits no items and terminates with an error.
func Thrown[T any](err error) Observable[T] {
	next := make(chan Item[T], 1)
	next <- Error[T](err)
	close(next)

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Timer returns an Observable that completes after a specified delay.
func Timer[T any](d Duration, opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := make(chan Item[T], 1)
	ctx := option.buildContext(emptyContext)

	go func() {
		defer close(next)
		select {
		case <-ctx.Done():
			return
		case <-time.After(d.duration()):
			return
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}
