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
	"container/ring"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/snivilised/lorax/enums"
)

// All determines whether all items emitted by an Observable meet some criteria.
func (o *ObservableImpl[T]) All(predicate Predicate[T], opts ...Option[T]) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &allOperator[T]{
			predicate: predicate,
			all:       true,
		}
	}, forceSeq, bypassGather, opts...)
}

type allOperator[T any] struct {
	predicate Predicate[T]
	all       bool
}

func (op *allOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if !op.predicate(item) {
		False[T]().SendContext(ctx, dst)

		op.all = false

		operatorOptions.stop()
	}
}

func (op *allOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *allOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.all {
		True[T]().SendContext(ctx, dst)
	}
}

func (op *allOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// !!! This might be an Opaque value representing the op-predicate
	//
	if value, err := TryBool(item); err == nil && !value {
		False[T]().SendContext(ctx, dst)

		op.all = false

		operatorOptions.stop()
	}
}

// Average calculates the average of numbers emitted by an Observable and emits the result.
func (o *ObservableImpl[T]) Average(opts ...Option[T],
) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	option := parseOptions(opts...)

	return single(o.parent, o, func() operator[T] {
		return &averageOperator[T]{
			calc: option.calc(),
		}
	}, forceSeq, bypassGather, opts...)
}

type averageOperator[T any] struct {
	sum   T
	count T
	calc  Calculator[T]
}

func (op *averageOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if !item.IsValue() || item.IsError() {
		Error[T](IllegalInputError{
			error: fmt.Sprintf("expected item value, got: %v (%v)", item.Desc(), item)},
		).SendContext(ctx, dst)
	}

	if op.calc == nil {
		Error[T](MissingCalcError{}).SendContext(ctx, dst)

		return
	}

	op.sum = op.calc.Add(op.sum, item.V)
	op.count = op.calc.Inc(op.count)
}

func (op *averageOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *averageOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.calc == nil {
		return
	}

	if op.calc.IsZero(op.count) {
		Of(op.calc.Zero()).SendContext(ctx, dst)
	} else {
		Of(op.calc.Div(op.sum, op.count)).SendContext(ctx, dst)
	}
}

func (op *averageOperator[T]) gatherNext(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	if v, err := TryOpaque[T, *averageOperator[T]](item); err != nil {
		op.sum = op.calc.Add(op.sum, v.sum)
		op.count = op.calc.Add(op.count, v.count)
	}
}

// BackOffRetry implements a backoff retry if a source Observable sends an error,
// resubscribe to it in the hopes that it will complete without error.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) BackOffRetry(backOffCfg backoff.BackOff,
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	f := func() error {
		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				close(next)

				return nil
			case i, ok := <-observe:
				if !ok {
					return nil
				}

				if i.IsError() {
					return i.E
				}

				i.SendContext(ctx, next)
			}
		}
	}

	go func() {
		if err := backoff.Retry(f, backOffCfg); err != nil {
			Error[T](err).SendContext(ctx, next)
			close(next)

			return
		}

		close(next)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Connect instructs a connectable Observable to begin emitting items to its subscribers.
func (o *ObservableImpl[T]) Connect(ctx context.Context) (context.Context, Disposable) {
	ctx, cancel := context.WithCancel(ctx)
	o.Observe(WithContext[T](ctx), connect[T]())

	return ctx, Disposable(cancel)
}

func (o *ObservableImpl[T]) Contains(equal Predicate[T], opts ...Option[T]) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &containsOperator[T]{
			equal:    equal,
			contains: false,
		}
	}, forceSeq, bypassGather, opts...)
}

type containsOperator[T any] struct {
	equal    Predicate[T]
	contains bool
}

func (op *containsOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.equal(item) {
		True[T]().SendContext(ctx, dst)

		op.contains = true

		operatorOptions.stop()
	}
}

func (op *containsOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *containsOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.contains {
		False[T]().SendContext(ctx, dst)
	}
}

func (op *containsOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if value, err := TryBool(item); err == nil && value {
		// we can reuse the item here instead of creating a new one
		// as we are sending true and we know that item already
		// represents true
		//
		item.SendContext(ctx, dst)
		operatorOptions.stop()

		op.contains = true
	}
}

// Count counts the number of items emitted by the source Observable and emit only this value.
func (o *ObservableImpl[T]) Count(opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &countOperator[T]{}
	}, forceSeq, bypassGather, opts...)
}

type countOperator[T any] struct {
	count int
}

func (op *countOperator[T]) next(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.count++
}

func (op *countOperator[T]) err(_ context.Context, _ Item[T],
	_ chan<- Item[T], operatorOptions operatorOptions[T],
) {
	operatorOptions.stop()
}

func (op *countOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	Num[T](op.count).SendContext(ctx, dst)
}

func (op *countOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// DefaultIfEmpty returns an Observable that emits the items emitted by the source
// Observable or a specified default item if the source Observable is empty.
func (o *ObservableImpl[T]) DefaultIfEmpty(defaultValue T, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &defaultIfEmptyOperator[T]{
			defaultValue: defaultValue,
			empty:        true,
		}
	}, forceSeq, bypassGather, opts...)
}

type defaultIfEmptyOperator[T any] struct {
	defaultValue T
	empty        bool
}

func (op *defaultIfEmptyOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	item.SendContext(ctx, dst)
}

func (op *defaultIfEmptyOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *defaultIfEmptyOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.empty {
		Of(op.defaultValue).SendContext(ctx, dst)
	}
}

func (op *defaultIfEmptyOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Distinct suppresses duplicate items in the original Observable and returns
// a new Observable.
func (o *ObservableImpl[T]) Distinct(apply Func[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &distinctOperator[T]{
			apply:  apply,
			keyset: make(map[interface{}]interface{}),
		}
	}, forceSeq, bypassGather, opts...)
}

type distinctOperator[T any] struct {
	apply  Func[T]
	keyset map[interface{}]interface{}
}

func (op *distinctOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	key, err := op.apply(ctx, item.V)
	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	_, ok := op.keyset[key]

	if !ok {
		item.SendContext(ctx, dst)
	}

	op.keyset[key] = nil
}

func (op *distinctOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *distinctOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *distinctOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if _, contains := op.keyset[item.V]; !contains {
		Of(item.V).SendContext(ctx, dst)

		op.keyset[item.V] = nil
	}
}

// DistinctUntilChanged suppresses consecutive duplicate items in the original Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) DistinctUntilChanged(apply Func[T], comparator Comparator[T],
	opts ...Option[T],
) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &distinctUntilChangedOperator[T]{
			apply:      apply,
			comparator: comparator,
		}
	}, forceSeq, bypassGather, opts...)
}

type distinctUntilChangedOperator[T any] struct {
	apply      Func[T]
	current    Item[T]
	comparator Comparator[T]
}

func (op *distinctUntilChangedOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T]) {
	key, err := op.apply(ctx, item.V)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	keyItem := Of(key)
	if op.comparator(op.current, keyItem) != 0 {
		item.SendContext(ctx, dst)

		op.current = keyItem
	}
}

func (op *distinctUntilChangedOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *distinctUntilChangedOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *distinctUntilChangedOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// DoOnCompleted registers a callback action that will be called once the
// Observable terminates.
func (o *ObservableImpl[T]) DoOnCompleted(completedFunc CompletedFunc,
	opts ...Option[T],
) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)
		defer completedFunc()

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-src:
				if !ok {
					return
				}

				if i.IsError() {
					return
				}
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// DoOnError registers a callback action that will be called if the
// Observable terminates abnormally.
func (o *ObservableImpl[T]) DoOnError(errFunc ErrFunc, opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-src:
				if !ok {
					return
				}

				if i.IsError() {
					errFunc(i.E)

					return
				}
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// DoOnNext registers a callback action that will be called on each item
// emitted by the Observable.
func (o *ObservableImpl[T]) DoOnNext(nextFunc NextFunc[T], opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-src:
				if !ok {
					return
				}

				if item.IsError() {
					return
				}

				nextFunc(item)
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// ElementAt emits only item n emitted by an Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) ElementAt(index uint, opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &elementAtOperator[T]{
			index: index,
		}
	}, forceSeq, bypassGather, opts...)
}

type elementAtOperator[T any] struct {
	index     uint
	takeCount int
	sent      bool
}

func (op *elementAtOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.takeCount == int(op.index) {
		item.SendContext(ctx, dst)

		op.sent = true

		operatorOptions.stop()

		return
	}

	op.takeCount++
}

func (op *elementAtOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *elementAtOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.sent {
		Error[T](&IllegalInputError{}).SendContext(ctx, dst)
	}
}

func (op *elementAtOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// Error returns the eventual Observable error.
// This method is blocking.
func (o *ObservableImpl[T]) Error(opts ...Option[T]) error {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	observe := o.iterable.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-observe:
			if !ok {
				return nil
			}

			if item.IsError() {
				return item.E
			}
		}
	}
}

// Errors returns an eventual list of Observable errors.
// This method is blocking
func (o *ObservableImpl[T]) Errors(opts ...Option[T]) []error {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	observe := o.iterable.Observe(opts...)
	errs := make([]error, 0)

	for {
		select {
		case <-ctx.Done():
			return []error{ctx.Err()}
		case item, ok := <-observe:
			if !ok {
				return errs
			}

			if item.IsError() {
				errs = append(errs, item.E)
			}
		}
	}
}

// Filter emits only those items from an Observable that pass a predicate test.
func (o *ObservableImpl[T]) Filter(apply Predicate[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return observable(o.parent, o, func() operator[T] {
		return &filterOperator[T]{apply: apply}
	}, forceSeq, bypassGather, opts...)
}

type filterOperator[T any] struct {
	apply Predicate[T]
}

func (op *filterOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if op.apply(item) {
		item.SendContext(ctx, dst)
	}
}

func (op *filterOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *filterOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *filterOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Find emits the first item passing a predicate then complete.
func (o *ObservableImpl[T]) Find(find Predicate[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = true
		bypassGather = true
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &findOperator[T]{
			find: find,
		}
	}, true, true, opts...)
}

type findOperator[T any] struct {
	find Predicate[T]
}

func (op *findOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.find(item) {
		item.SendContext(ctx, dst)
		operatorOptions.stop()
	}
}

func (op *findOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *findOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *findOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// First returns new Observable which emit only first item.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) First(opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &firstOperator[T]{}
	}, forceSeq, bypassGather, opts...)
}

type firstOperator[T any] struct{}

func (op *firstOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	item.SendContext(ctx, dst)
	operatorOptions.stop()
}

func (op *firstOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *firstOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *firstOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// FirstOrDefault returns new Observable which emit only first item.
// If the observable fails to emit any items, it emits a default value.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) FirstOrDefault(defaultValue T, opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &firstOrDefaultOperator[T]{
			defaultValue: defaultValue,
		}
	}, forceSeq, bypassGather, opts...)
}

type firstOrDefaultOperator[T any] struct {
	defaultValue T
	sent         bool
}

func (op *firstOrDefaultOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	item.SendContext(ctx, dst)

	op.sent = true

	operatorOptions.stop()
}

func (op *firstOrDefaultOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *firstOrDefaultOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.sent {
		Of(op.defaultValue).SendContext(ctx, dst)
	}
}

func (op *firstOrDefaultOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// FlatMap transforms the items emitted by an Observable into Observables,
// then flatten the emissions from those into a single Observable.
func (o *ObservableImpl[T]) FlatMap(apply ItemToObservable[T],
	opts ...Option[T],
) Observable[T] {
	f := func(ctx context.Context, next chan Item[T], option Option[T], opts ...Option[T]) {
		defer close(next)

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return

			case item, ok := <-observe:
				if !ok {
					return
				}

				observe2 := apply(item).Observe(opts...)

			loop2:
				for {
					select {
					case <-ctx.Done():
						return

					case item, ok := <-observe2:
						if !ok {
							break loop2
						}

						if item.IsError() {
							item.SendContext(ctx, next)
							if option.getErrorStrategy() == enums.StopOnError {
								return
							}
						} else if !item.SendContext(ctx, next) {
							return
						}
					}
				}
			}
		}
	}

	return customObservableOperator(o.parent, f, opts...)
}

// ForEach subscribes to the Observable and receives notifications for each element.
func (o *ObservableImpl[T]) ForEach(nextFunc NextFunc[T],
	errFunc ErrFunc, completedFunc CompletedFunc, opts ...Option[T],
) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)

		for {
			select {
			case <-ctx.Done():
				completedFunc()

				return

			case item, ok := <-src:
				if !ok {
					completedFunc()

					return
				}

				if item.IsError() {
					errFunc(item.E)

					break
				}

				nextFunc(item)
			}
		}
	}

	ctx := o.parent

	if ctx == nil {
		ctx = context.Background()
	}

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// GroupBy divides an Observable into a set of Observables that
// each emit a different group of items from the original Observable,
// organized by key.
func (o *ObservableImpl[T]) GroupBy(length int,
	distribution DistributionFunc[T], opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	s := make([]Item[T], length)
	chs := make([]chan Item[T], length)

	for i := 0; i < length; i++ {
		ch := option.buildChannel()
		chs[i] = ch

		s[i] = Opaque[T](&ObservableImpl[T]{
			iterable: newChannelIterable(ch),
		})
	}

	go func() {
		observe := o.Observe(opts...)

		defer func() {
			for i := 0; i < length; i++ {
				close(chs[i])
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case item, ok := <-observe:
				if !ok {
					return
				}

				// Is this where we receive the Opaque *ObservableImpl item?
				//
				idx := distribution(item)
				if idx >= length {
					err := Error[T](IndexOutOfBoundError{
						error: fmt.Sprintf("index %d, length %d", idx, length),
					})
					for i := 0; i < length; i++ {
						err.SendContext(ctx, chs[i])
					}

					return
				}

				item.SendContext(ctx, chs[idx])
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newSliceIterable(s, opts...),
	}
}

// GroupedObservable is the observable type emitted by the GroupByDynamic operator.
type GroupedObservable[T any] struct {
	Observable[T]
	// Key is the distribution key
	Key string
}

// GroupByDynamic divides an Observable into a dynamic set of
// Observables that each emit GroupedObservable from the original
// Observable, organized by key.
func (o *ObservableImpl[T]) GroupByDynamic(distribution DynamicDistributionFunc[T],
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	chs := make(map[string]chan Item[T])

	go func() {
		observe := o.Observe(opts...)
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case item, ok := <-observe:
				if !ok {
					break loop
				}

				idx := distribution(item)
				ch, contains := chs[idx]

				if !contains {
					ch = option.buildChannel()
					chs[idx] = ch
					Opaque[T](GroupedObservable[T]{ // TODO: where is this received?
						Observable: &ObservableImpl[T]{
							iterable: newChannelIterable(ch),
						},
						Key: idx,
					}).SendContext(ctx, next)
				}
				item.SendContext(ctx, ch)
			}
		}

		for _, ch := range chs {
			close(ch)
		}

		close(next)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// IgnoreElements ignores all items emitted by the source ObservableSource except
// for the errors. Cannot be run in parallel.
func (o *ObservableImpl[T]) IgnoreElements(opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &ignoreElementsOperator[T]{}
	}, forceSeq, bypassGather, opts...)
}

type ignoreElementsOperator[T any] struct{}

func (op *ignoreElementsOperator[T]) next(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

func (op *ignoreElementsOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T]) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *ignoreElementsOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *ignoreElementsOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Last returns a new Observable which emit only last item.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) Last(opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &lastOperator[T]{
			empty: true,
		}
	}, forceSeq, bypassGather, opts...)
}

type lastOperator[T any] struct {
	last  Item[T]
	empty bool
}

func (op *lastOperator[T]) next(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.last = item
	op.empty = false
}

func (op *lastOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T]) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *lastOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		op.last.SendContext(ctx, dst)
	}
}

func (op *lastOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

func (o *ObservableImpl[T]) LastOrDefault(defaultValue T, opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &lastOrDefaultOperator[T]{
			defaultValue: defaultValue,
			empty:        true,
		}
	}, forceSeq, bypassGather, opts...)
}

type lastOrDefaultOperator[T any] struct {
	defaultValue T
	last         Item[T]
	empty        bool
}

func (op *lastOrDefaultOperator[T]) next(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.last = item
	op.empty = false
}

func (op *lastOrDefaultOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *lastOrDefaultOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		op.last.SendContext(ctx, dst)
	} else {
		Of(op.defaultValue).SendContext(ctx, dst)
	}
}

func (op *lastOrDefaultOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Max determines and emits the maximum-valued item emitted by an Observable
// according to a comparator.
func (o *ObservableImpl[T]) Max(comparator Comparator[T], initLimit InitLimit[T],
	opts ...Option[T],
) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &maxOperator[T]{
			comparator: comparator,
			max:        initLimit(),
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

type maxOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	max        Item[T]
}

func (op *maxOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	if op.comparator(op.max, item) < 0 {
		op.max = item
	}
}

func (op *maxOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *maxOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		// Using Num here instead of Of, means that the Max operator only works
		// for numbers and not any other Item type, which is probably not what
		// we wanted. We initially tried to limit the Max operator to number,
		// but in the fullness of time, this looks to be incorrect. (FuncN!!!)
		// If we had a channel of widgets, this wouldn't work when it supposed to be
		// able to, so we need a new design. But this was a valuable learning
		// experience. We need to radically redesign Min/Max/Map operators to
		// be able to work properly, and also can detect when a value has been
		// set. In the legacy rxgo, it depended on reflection and the ability
		// to perform a test like:
		//
		// if op.max == nil {
		// 	op.max = item.V
		// } else {
		// 	if op.comparator(op.max, item.V) < 0 {
		// 		op.max = item.V
		// 	}
		// }
		//
		// There should be no internal code that tries to compare values, because
		// when generics are in play, only the client knows how to do this, so
		// there should be a way for the client to implement these types of checks
		// themselves, probably by passing in a new function like comparator.
		//
		op.max.SendContext(ctx, dst)
	}
}

func (op *maxOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if result, err := TryOpaque[T, *maxOperator[T]](item); err == nil {
		// this is sent by runParallel
		//
		op.next(ctx, result.max, dst, operatorOptions)
	}
}

// Min determines and emits the minimum-valued item emitted by an Observable
// according to a comparator.
func (o *ObservableImpl[T]) Min(comparator Comparator[T], initLimit InitLimit[T],
	opts ...Option[T],
) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &minOperator[T]{
			min:        initLimit(),
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

type minOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	min        Item[T]
	limit      func(value T) bool
}

func (op *minOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	if op.comparator(op.min, item) > 0 {
		op.min = item
	}
}

func (op *minOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *minOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		op.min.SendContext(ctx, dst)
	}
}

func (op *minOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if result, err := TryOpaque[T, *minOperator[T]](item); err == nil {
		// this is sent by runParallel
		//
		op.next(ctx, result.min, dst, operatorOptions)
	}
}

// Map transforms the items emitted by an Observable by applying a function to each item.
func (o *ObservableImpl[T]) Map(apply Func[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return observable(o.parent, o, func() operator[T] {
		return &mapOperator[T]{
			apply: apply,
		}
	}, forceSeq, bypassGather, opts...)
}

type mapOperator[T any] struct {
	apply Func[T]
}

func (op *mapOperator[T]) next(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	res, err := op.apply(ctx, item.V)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	Of(res).SendContext(ctx, dst)
}

func (op *mapOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *mapOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], _ operatorOptions[T],
) {
	item.SendContext(ctx, dst)
}

func (o *ObservableImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return o.iterable.Observe(opts...)
}

// OnErrorResumeNext instructs an Observable to pass control to another Observable rather than invoking
// onError if it encounters an error.
func (o *ObservableImpl[T]) OnErrorResumeNext(resumeSequence ErrorToObservable[T],
	opts ...Option[T],
) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &onErrorResumeNextOperator[T]{
			resumeSequence: resumeSequence,
		}
	}, forceSeq, bypassGather, opts...)
}

type onErrorResumeNextOperator[T any] struct {
	resumeSequence ErrorToObservable[T]
}

func (op *onErrorResumeNextOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T]) {
	item.SendContext(ctx, dst)
}

func (op *onErrorResumeNextOperator[T]) err(_ context.Context, item Item[T],
	_ chan<- Item[T], operatorOptions operatorOptions[T],
) {
	operatorOptions.resetIterable(op.resumeSequence(item.E))
}

func (op *onErrorResumeNextOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *onErrorResumeNextOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

func (o *ObservableImpl[T]) OnErrorReturn(resumeFunc ErrorFunc[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &onErrorReturnOperator[T]{
			resumeFunc: resumeFunc,
		}
	}, forceSeq, bypassGather, opts...)
}

type onErrorReturnOperator[T any] struct {
	resumeFunc ErrorFunc[T]
}

func (op *onErrorReturnOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T]) {
	item.SendContext(ctx, dst)
}

func (op *onErrorReturnOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	Of[T](op.resumeFunc(item.E)).SendContext(ctx, dst)
}

func (op *onErrorReturnOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *onErrorReturnOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

func (o *ObservableImpl[T]) OnErrorReturnItem(resume T, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &onErrorReturnItemOperator[T]{
			resume: resume,
		}
	}, true, false, opts...)
}

type onErrorReturnItemOperator[T any] struct {
	resume T
}

func (op *onErrorReturnItemOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	item.SendContext(ctx, dst)
}

func (op *onErrorReturnItemOperator[T]) err(ctx context.Context, _ Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	Of(op.resume).SendContext(ctx, dst)
}

func (op *onErrorReturnItemOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *onErrorReturnItemOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Reduce applies a function to each item emitted by an Observable, sequentially,
// and emit the final value.
func (o *ObservableImpl[T]) Reduce(apply Func2[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &reduceOperator[T]{
			apply: apply,
			acc:   Zero[T](),
			empty: true,
		}
	}, forceSeq, bypassGather, opts...)
}

type reduceOperator[T any] struct {
	apply Func2[T]
	acc   Item[T]
	empty bool
}

func (op *reduceOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	op.empty = false
	v, err := op.apply(ctx, op.acc, item)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		op.empty = true

		return
	}

	op.acc.V = v
}

func (op *reduceOperator[T]) err(_ context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	dst <- item

	op.empty = true

	operatorOptions.stop()
}

func (op *reduceOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		op.acc.SendContext(ctx, dst)
	}
}

func (op *reduceOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if result, err := TryOpaque[T, *reduceOperator[T]](item); err == nil {
		op.next(ctx, result.acc, dst, operatorOptions)
	}
}

// Repeat returns an Observable that repeats the sequence of items emitted
// by the source Observable at most count times, at a particular frequency.
// Cannot run in parallel.
func (o *ObservableImpl[T]) Repeat(count int64, frequency Duration, opts ...Option[T]) Observable[T] {
	if count != Infinite {
		if count < 0 {
			return Thrown[T](IllegalInputError{error: "count must be positive"})
		}
	}

	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &repeatOperator[T]{
			count:     count,
			frequency: frequency,
			seq:       make([]Item[T], 0),
		}
	}, forceSeq, bypassGather, opts...)
}

type repeatOperator[T any] struct {
	count     int64
	frequency Duration
	seq       []Item[T]
}

func (op *repeatOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	item.SendContext(ctx, dst)
	op.seq = append(op.seq, item)
}

func (op *repeatOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *repeatOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	for {
		select {
		default:
		case <-ctx.Done():
			return
		}

		if op.count != Infinite {
			if op.count == 0 {
				break
			}
		}

		if op.frequency != nil {
			time.Sleep(op.frequency.duration())
		}

		for _, v := range op.seq {
			v.SendContext(ctx, dst)
		}

		op.count--
	}
}

func (op *repeatOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Retry retries if a source Observable sends an error, resubscribe to
// it in the hopes that it will complete without error. Cannot be run in parallel.
func (o *ObservableImpl[T]) Retry(count int, shouldRetry ShouldRetryFunc,
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)

	go func() {
		observe := o.Observe(opts...)
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case i, ok := <-observe:
				if !ok {
					break loop
				}

				if i.IsError() {
					count--

					if count < 0 || !shouldRetry(i.E) {
						i.SendContext(ctx, next)
						break loop
					}

					observe = o.Observe(opts...)
				} else {
					i.SendContext(ctx, next)
				}
			}
		}
		close(next)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Run creates an Observer without consuming the emitted items.
func (o *ObservableImpl[T]) Run(opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go func() {
		defer close(dispose)

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-observe:
				if !ok {
					return
				}
			}
		}
	}()

	return dispose
}

// Sample returns an Observable that emits the most recent items emitted by the source
// Iterable whenever the input Iterable emits an item.
func (o *ObservableImpl[T]) Sample(iterable Iterable[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	itCh := make(chan Item[T])
	obsCh := make(chan Item[T])

	go func() {
		defer close(obsCh)

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-observe:
				if !ok {
					return
				}

				i.SendContext(ctx, obsCh)
			}
		}
	}()

	go func() {
		defer close(itCh)

		observe := iterable.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-observe:
				if !ok {
					return
				}

				i.SendContext(ctx, itCh)
			}
		}
	}()

	go func() {
		defer close(next)

		var lastEmittedItem Item[T]

		isItemWaitingToBeEmitted := false

		for {
			select {
			case _, ok := <-itCh:
				if ok {
					if isItemWaitingToBeEmitted {
						next <- lastEmittedItem

						isItemWaitingToBeEmitted = false
					}
				} else {
					return
				}
			case item, ok := <-obsCh:
				if ok {
					lastEmittedItem = item
					isItemWaitingToBeEmitted = true
				} else {
					return
				}
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Scan apply a Func2 to each item emitted by an Observable, sequentially, and
// emit each successive value. Cannot be run in parallel.
func (o *ObservableImpl[T]) Scan(apply Func2[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &scanOperator[T]{
			apply: apply,
		}
	}, forceSeq, bypassGather, opts...)
}

type scanOperator[T any] struct {
	apply   Func2[T]
	current Item[T]
}

func (op *scanOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	v, err := op.apply(ctx, op.current, item)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	op.current = Of(v)
	op.current.SendContext(ctx, dst)
}

func (op *scanOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *scanOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *scanOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Compares first items of two sequences and returns true if they are equal and false if
// they are not. Besides, it returns two new sequences - input sequences without compared items.
func popAndCompareFirstItems[T any]( //nolint:gocritic // foo
	inputSequence1 []Item[T],
	inputSequence2 []Item[T],
	comparator Comparator[T],
) (bool, []Item[T], []Item[T]) {
	if len(inputSequence1) > 0 && len(inputSequence2) > 0 {
		s1, sequence1 := inputSequence1[0], inputSequence1[1:]
		s2, sequence2 := inputSequence2[0], inputSequence2[1:]

		return comparator(s1, s2) == 0, sequence1, sequence2
	}

	return true, inputSequence1, inputSequence2
}

// Send sends the items to a given channel.
func (o *ObservableImpl[T]) Send(output chan<- Item[T], opts ...Option[T]) {
	go func() {
		option := parseOptions(opts...)
		ctx := option.buildContext(o.parent)
		observe := o.Observe(opts...)
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case i, ok := <-observe:
				if !ok {
					break loop
				}

				if i.IsError() {
					output <- i
					break loop
				}

				i.SendContext(ctx, output)
			}
		}
		close(output)
	}()
}

// SequenceEqual emits true if an Observable and the input Observable emit the same items,
// in the same order, with the same termination state. Otherwise, it emits false.
func (o *ObservableImpl[T]) SequenceEqual(iterable Iterable[T],
	comparator Comparator[T],
	opts ...Option[T],
) Single[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	itCh := make(chan Item[T])
	obsCh := make(chan Item[T])

	go func() {
		defer close(obsCh)

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-observe:
				if !ok {
					return
				}

				i.SendContext(ctx, obsCh)
			}
		}
	}()

	go func() {
		defer close(itCh)

		observe := iterable.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-observe:
				if !ok {
					return
				}

				i.SendContext(ctx, itCh)
			}
		}
	}()

	go func() {
		var (
			mainSequence, obsSequence []Item[T]
		)

		areCorrect := true
		isMainChannelClosed := false
		isObsChannelClosed := false

		for {
			select {
			case item, ok := <-itCh:
				if ok {
					mainSequence = append(mainSequence, item)

					areCorrect, mainSequence, obsSequence = popAndCompareFirstItems(
						mainSequence, obsSequence,
						comparator,
					)
				} else {
					isMainChannelClosed = true
				}

			case item, ok := <-obsCh:
				if ok {
					obsSequence = append(obsSequence, item)
					areCorrect, mainSequence, obsSequence = popAndCompareFirstItems(
						mainSequence, obsSequence,
						comparator,
					)
				} else {
					isObsChannelClosed = true
				}
			}

			if !areCorrect || (isMainChannelClosed && isObsChannelClosed) {
				break
			}
		}

		Bool[T](
			areCorrect && len(mainSequence) == 0 && len(obsSequence) == 0,
		).SendContext(ctx, next)

		close(next)
	}()

	return &SingleImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Serialize forces an Observable to make serialized calls and to be well-behaved.
func (o *ObservableImpl[T]) Serialize(from int, identifier func(any) int,
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	minHeap := binaryheap.NewWith(func(a, b any) int {
		return a.(int) - b.(int)
	})
	counter := int64(from)
	items := make(map[int]T)

	go func() {
		src := o.Observe(opts...)

		defer close(next)

		for {
			select {
			case <-ctx.Done():
				return

			case item, ok := <-src:
				if !ok {
					return
				}

				if item.IsError() {
					next <- item
					return
				}

				id := identifier(item.V)
				minHeap.Push(id)

				items[id] = item.V

				for !minHeap.Empty() {
					v, _ := minHeap.Peek()
					id := v.(int) //nolint:errcheck // TODO

					if atomic.LoadInt64(&counter) == int64(id) {
						if itemValue, contains := items[id]; contains {
							minHeap.Pop()
							delete(items, id)
							Of(itemValue).SendContext(ctx, next)

							counter++

							continue
						}
					}

					break
				}
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Skip suppresses the first n items in the original Observable and
// returns a new Observable with the rest items.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) Skip(nth uint, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &skipOperator[T]{
			nth: nth,
		}
	}, forceSeq, bypassGather, opts...)
}

type skipOperator[T any] struct {
	nth       uint
	skipCount int
}

func (op *skipOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if op.skipCount < int(op.nth) {
		op.skipCount++
		return
	}

	item.SendContext(ctx, dst)
}

func (op *skipOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *skipOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *skipOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// SkipLast suppresses the last n items in the original Observable and
// returns a new Observable with the rest items.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) SkipLast(nth uint, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &skipLastOperator[T]{
			nth: nth,
		}
	}, forceSeq, bypassGather, opts...)
}

type skipLastOperator[T any] struct {
	nth       uint
	skipCount int
}

func (op *skipLastOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.skipCount >= int(op.nth) {
		operatorOptions.stop()
		return
	}

	op.skipCount++

	item.SendContext(ctx, dst)
}

func (op *skipLastOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *skipLastOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *skipLastOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// SkipWhile discard items emitted by an Observable until a specified condition becomes false.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) SkipWhile(apply Predicate[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &skipWhileOperator[T]{
			apply: apply,
			skip:  true,
		}
	}, forceSeq, bypassGather, opts...)
}

type skipWhileOperator[T any] struct {
	apply Predicate[T]
	skip  bool
}

func (op *skipWhileOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if !op.skip {
		item.SendContext(ctx, dst)
	} else if !op.apply(item) {
		op.skip = false

		item.SendContext(ctx, dst)
	}
}

func (op *skipWhileOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *skipWhileOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *skipWhileOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// StartWith emits a specified Iterable before beginning to emit the items from the source Observable.
func (o *ObservableImpl[T]) StartWith(iterable Iterable[T], opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)

	go func() {
		defer close(next)

		observe := iterable.Observe(opts...)
	loop1:
		for {
			select {
			case <-ctx.Done():
				break loop1
			case i, ok := <-observe:
				if !ok {
					break loop1
				}
				if i.IsError() {
					next <- i
					return
				}
				i.SendContext(ctx, next)
			}
		}

		observe = o.Observe(opts...)

	loop2:
		for {
			select {
			case <-ctx.Done():
				break loop2
			case i, ok := <-observe:
				if !ok {
					break loop2
				}

				if i.IsError() {
					i.SendContext(ctx, next)
					return
				}
				i.SendContext(ctx, next)
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Sum calculates the average emitted by an Observable and emits the result
func (o *ObservableImpl[T]) Sum(opts ...Option[T]) OptionalSingle[T] {
	options := parseOptions(opts...)
	calc := options.calc()

	return o.Reduce(func(_ context.Context, acc, item Item[T]) (T, error) {
		if calc == nil {
			var (
				zero T
			)

			return zero, MissingCalcError{}
		}

		return calc.Add(acc.V, item.V), nil
	}, opts...)
}

// Take emits only the first n items emitted by an Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) Take(nth uint, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &takeOperator[T]{
			nth: nth,
		}
	}, forceSeq, bypassGather, opts...)
}

type takeOperator[T any] struct {
	nth       uint
	takeCount int
}

func (op *takeOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.takeCount >= int(op.nth) {
		operatorOptions.stop()

		return
	}

	op.takeCount++

	item.SendContext(ctx, dst)
}

func (op *takeOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *takeOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *takeOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// TakeLast emits only the last n items emitted by an Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) TakeLast(nth uint, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		n := int(nth)

		return &takeLast[T]{
			n: n,
			r: ring.New(n),
		}
	}, forceSeq, bypassGather, opts...)
}

type takeLast[T any] struct {
	n     int
	r     *ring.Ring
	count int
}

func (op *takeLast[T]) next(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.count++
	op.r.Value = item.V
	op.r = op.r.Next()
}

func (op *takeLast[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *takeLast[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.count < op.n {
		remaining := op.n - op.count

		if remaining <= op.count {
			op.r = op.r.Move(op.n - op.count)
		} else {
			op.r = op.r.Move(-op.count)
		}

		op.n = op.count
	}

	for i := 0; i < op.n; i++ {
		Of(op.r.Value.(T)).SendContext(ctx, dst)
		op.r = op.r.Next()
	}
}

func (op *takeLast[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// TakeUntil returns an Observable that emits items emitted by the source Observable,
// checks the specified predicate for each item, and then completes when the condition is satisfied.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) TakeUntil(apply Predicate[T],
	opts ...Option[T],
) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &takeUntilOperator[T]{
			apply: apply,
		}
	}, forceSeq, bypassGather, opts...)
}

type takeUntilOperator[T any] struct {
	apply Predicate[T]
}

func (op *takeUntilOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	item.SendContext(ctx, dst)

	if op.apply(item) {
		operatorOptions.stop()

		return
	}
}

func (op *takeUntilOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *takeUntilOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *takeUntilOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// TakeWhile returns an Observable that emits items emitted by the source ObservableSource so long as each
// item satisfied a specified condition, and then completes as soon as this condition is not satisfied.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) TakeWhile(apply Predicate[T],
	opts ...Option[T],
) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &takeWhileOperator[T]{
			apply: apply,
		}
	}, forceSeq, bypassGather, opts...)
}

type takeWhileOperator[T any] struct {
	apply Predicate[T]
}

func (op *takeWhileOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if !op.apply(item) {
		operatorOptions.stop()

		return
	}

	item.SendContext(ctx, dst)
}

func (op *takeWhileOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *takeWhileOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *takeWhileOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

func (o *ObservableImpl[T]) TimeInterval(opts ...Option[T]) Observable[T] {
	f := func(ctx context.Context, next chan Item[T], option Option[T], opts ...Option[T]) {
		defer close(next)

		observe := o.Observe(opts...)
		latest := time.Now().UTC()

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-observe:
				if !ok {
					return
				}

				if item.IsError() {
					if !item.SendContext(ctx, next) {
						return
					}

					if option.getErrorStrategy() == enums.StopOnError {
						return
					}
				} else {
					now := time.Now().UTC()

					if !Opaque[T](now.Sub(latest)).SendContext(ctx, next) {
						return
					}

					latest = now
				}
			}
		}
	}

	return customObservableOperator(o.parent, f, opts...)
}

func (o *ObservableImpl[T]) Timestamp(opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &timestampOperator[T]{}
	}, forceSeq, bypassGather, opts...)
}

type timestampOperator[T any] struct {
}

func (op *timestampOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	Opaque[T](&TimestampItem[T]{
		Timestamp: time.Now().UTC(),
		V:         item.V,
	}).SendContext(ctx, dst)
}

func (op *timestampOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *timestampOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *timestampOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// ToSlice collects all items from an Observable and emit them in a slice and
// an optional error. Cannot be run in parallel.
func (o *ObservableImpl[T]) ToSlice(initialCapacity int, opts ...Option[T]) ([]Item[T], error) {
	const (
		forceSeq     = true
		bypassGather = false
	)

	op := &toSliceOperator[T]{
		s: make([]Item[T], 0, initialCapacity),
	}

	<-observable(o.parent, o, func() operator[T] {
		return op
	}, forceSeq, bypassGather, opts...).Run()

	return op.s, op.observableErr
}

type toSliceOperator[T any] struct {
	s             []Item[T]
	observableErr error
}

func (op *toSliceOperator[T]) next(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.s = append(op.s, item)
}

func (op *toSliceOperator[T]) err(_ context.Context, item Item[T],
	_ chan<- Item[T], operatorOptions operatorOptions[T]) {
	op.observableErr = item.E

	operatorOptions.stop()
}

func (op *toSliceOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *toSliceOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// WindowWithCount periodically subdivides items from an Observable into Observable windows of a given size and emit these windows
// rather than emitting the items one at a time.
func (o *ObservableImpl[T]) WindowWithCount(count int, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	if count < 0 {
		return Thrown[T](IllegalInputError{
			error: "count must be positive or nil",
		})
	}

	option := parseOptions(opts...)

	return observable(o.parent, o, func() operator[T] {
		return &windowWithCountOperator[T]{
			count:  count,
			option: option,
		}
	}, forceSeq, bypassGather, opts...)
}

type windowWithCountOperator[T any] struct {
	count          int
	iCount         int
	currentChannel chan Item[T]
	option         Option[T]
}

func (op *windowWithCountOperator[T]) pre(ctx context.Context, dst chan<- Item[T]) {
	if op.currentChannel == nil {
		ch := op.option.buildChannel()
		op.currentChannel = ch

		Opaque[T](FromChannel(ch)).SendContext(ctx, dst)
	}
}

func (op *windowWithCountOperator[T]) post(ctx context.Context, dst chan<- Item[T]) {
	if op.iCount == op.count {
		op.iCount = 0
		close(op.currentChannel)

		ch := op.option.buildChannel()
		op.currentChannel = ch
		Opaque[T](FromChannel(ch)).SendContext(ctx, dst)
	}
}

func (op *windowWithCountOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	op.pre(ctx, dst)
	op.currentChannel <- item

	op.iCount++

	op.post(ctx, dst)
}

func (op *windowWithCountOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T]) {
	op.pre(ctx, dst)

	op.currentChannel <- item

	op.iCount++

	op.post(ctx, dst)
	operatorOptions.stop()
}

func (op *windowWithCountOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
	if op.currentChannel != nil {
		close(op.currentChannel)
	}
}

func (op *windowWithCountOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// WindowWithTime periodically subdivides items from an Observable into Observables based on timed windows
// and emit them rather than emitting the items one at a time.
func (o *ObservableImpl[T]) WindowWithTime(timespan Duration, opts ...Option[T]) Observable[T] {
	if timespan == nil {
		return Thrown[T](IllegalInputError{
			error: "timespan must no be nil",
		})
	}

	f := func(ctx context.Context, next chan Item[T], option Option[T], opts ...Option[T]) {
		observe := o.Observe(opts...)
		ch := option.buildChannel()
		done := make(chan struct{})
		empty := true
		mutex := sync.Mutex{}

		if !Opaque[T](FromChannel(ch)).SendContext(ctx, next) {
			return
		}

		go func() {
			defer func() {
				mutex.Lock()
				close(ch)
				mutex.Unlock()
			}()
			defer close(next)

			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				case <-time.After(timespan.duration()):
					mutex.Lock()

					if empty {
						mutex.Unlock()
						continue
					}

					close(ch)

					empty = true

					ch = option.buildChannel()

					if !Opaque[T](FromChannel(ch)).SendContext(ctx, next) {
						close(done)

						return
					}
					mutex.Unlock()
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case item, ok := <-observe:
				if !ok {
					close(done)

					return
				}

				if item.IsError() {
					mutex.Lock()

					if !item.SendContext(ctx, ch) {
						mutex.Unlock()
						close(done)

						return
					}
					mutex.Unlock()

					if option.getErrorStrategy() == enums.StopOnError {
						close(done)

						return
					}
				}

				mutex.Lock()
				if !item.SendContext(ctx, ch) {
					mutex.Unlock()

					return
				}

				empty = false

				mutex.Unlock()
			}
		}
	}

	return customObservableOperator(o.parent, f, opts...)
}

// WindowWithTimeOrCount periodically subdivides items from an Observable into Observables based on timed windows or a specific size
// and emit them rather than emitting the items one at a time.
func (o *ObservableImpl[T]) WindowWithTimeOrCount(timespan Duration,
	count int, opts ...Option[T],
) Observable[T] {
	if timespan == nil {
		return Thrown[T](IllegalInputError{
			error: "timespan must not be nil",
		})
	}

	if count < 0 {
		return Thrown[T](IllegalInputError{
			error: "count must be positive or nil",
		})
	}

	f := func(ctx context.Context, next chan Item[T], option Option[T],
		opts ...Option[T],
	) {
		observe := o.Observe(opts...)
		ch := option.buildChannel()
		done := make(chan struct{})
		mutex := sync.Mutex{}
		iCount := 0

		if !Opaque[T](FromChannel(ch)).SendContext(ctx, next) {
			return
		}

		go func() {
			defer func() {
				mutex.Lock()
				close(ch)
				mutex.Unlock()
			}()
			defer close(next)

			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				case <-time.After(timespan.duration()):
					mutex.Lock()
					if iCount == 0 {
						mutex.Unlock()

						continue
					}

					close(ch)

					iCount = 0
					ch = option.buildChannel()

					if !Opaque[T](FromChannel(ch)).SendContext(ctx, next) {
						close(done)

						return
					}
					mutex.Unlock()
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case item, ok := <-observe:
				if !ok {
					close(done)

					return
				}

				if item.IsError() {
					mutex.Lock()
					if !item.SendContext(ctx, ch) {
						mutex.Unlock()

						close(done)

						return
					}
					mutex.Unlock()

					if option.getErrorStrategy() == enums.StopOnError {
						close(done)

						return
					}
				}

				mutex.Lock()

				if !item.SendContext(ctx, ch) {
					mutex.Unlock()

					return
				}

				iCount++

				if iCount == count {
					close(ch)

					iCount = 0
					ch = option.buildChannel()

					if !Opaque[T](FromChannel(ch)).SendContext(ctx, next) {
						mutex.Unlock()
						close(done)

						return
					}
				}
				mutex.Unlock()
			}
		}
	}

	return customObservableOperator(o.parent, f, opts...)
}

// ZipFromIterable merges the emissions of an Iterable via a specified function
// and emit single items for each combination based on the results of this function.
func (o *ObservableImpl[T]) ZipFromIterable(iterable Iterable[T], zipper Func2[T],
	opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)

	go func() {
		defer close(next)

		it1 := o.Observe(opts...)
		it2 := iterable.Observe(opts...)
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case i1, ok := <-it1:
				if !ok {
					break loop
				}

				if i1.IsError() {
					i1.SendContext(ctx, next)

					return
				}

				for {
					select {
					case <-ctx.Done():
						break loop
					case i2, ok := <-it2:
						if !ok {
							break loop
						}

						if i2.IsError() {
							i2.SendContext(ctx, next)
							return
						}

						v, err := zipper(ctx, i1, i2)

						if err != nil {
							Error[T](err).SendContext(ctx, next)

							return
						}

						Of(v).SendContext(ctx, next)
						continue loop
					}
				}
			}
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}
