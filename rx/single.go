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

import "context"

// Single is a observable with a single element.
type Single[T any] interface {
	Iterable[T]
	Filter(apply Predicate[T], opts ...Option[T]) OptionalSingle[T]
	Get(opts ...Option[T]) (Item[T], error)
	Map(apply Func[T], opts ...Option[T]) Single[T]
	Run(opts ...Option[T]) Disposed
}

// SingleImpl implements Single.
type SingleImpl[T any] struct {
	parent   context.Context
	iterable Iterable[T]
}

// Filter emits only those items from an Observable that pass a predicate test.
func (s *SingleImpl[T]) Filter(apply Predicate[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = true
		bypassGather = true
	)

	return optionalSingle(s.parent, s, func() operator[T] {
		return &filterOperatorSingle[T]{apply: apply}
	}, forceSeq, bypassGather, opts...)
}

// Get returns the item. The error returned is if the context has been cancelled.
// This method is blocking.
func (s *SingleImpl[T]) Get(opts ...Option[T]) (Item[T], error) {
	option := parseOptions(opts...)
	ctx := option.buildContext(s.parent)

	observe := s.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return Item[T]{}, ctx.Err()
		case v := <-observe:
			return v, nil
		}
	}
}

// Map transforms the items emitted by a Single by applying a function to each item.
func (s *SingleImpl[T]) Map(apply Func[T], opts ...Option[T]) Single[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return single(s.parent, s, func() operator[T] {
		return &mapOperatorSingle[T]{apply: apply}
	}, forceSeq, bypassGather, opts...)
}

type mapOperatorSingle[T any] struct {
	apply Func[T]
}

func (op *mapOperatorSingle[T]) next(ctx context.Context,
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

func (op *mapOperatorSingle[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperatorSingle[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *mapOperatorSingle[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], _ operatorOptions[T],
) {
	// TODO: switch item.V.(type) {
	// case *mapOperatorSingle:
	// 	return
	// }
	//
	item.SendContext(ctx, dst)
	panic("mapOperatorSingle.gatherNext:NOT-IMPL")
}

// Observe observes a Single by returning its channel.
func (s *SingleImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return s.iterable.Observe(opts...)
}

type filterOperatorSingle[T any] struct {
	apply Predicate[T]
}

func (op *filterOperatorSingle[T]) next(ctx context.Context,
	item Item[T], dst chan<- Item[T], _ operatorOptions[T],
) {
	if op.apply(item) {
		item.SendContext(ctx, dst)
	}
}

func (op *filterOperatorSingle[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *filterOperatorSingle[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *filterOperatorSingle[T]) gatherNext(_ context.Context,
	_ Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
}

// Run creates an observer without consuming the emitted items.
func (s *SingleImpl[T]) Run(opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	option := parseOptions(opts...)
	ctx := option.buildContext(s.parent)

	go func() {
		defer close(dispose)

		observe := s.Observe(opts...)

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
