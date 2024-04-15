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
)

// OptionalSingle is an optional single.
type OptionalSingle[T any] interface {
	Iterable[T]
	Get(opts ...Option[T]) (Item[T], error)
	Map(apply Func[T], opts ...Option[T]) OptionalSingle[T]
	Run(opts ...Option[T]) Disposed
}

// OptionalSingleImpl implements OptionalSingle.
type OptionalSingleImpl[T any] struct {
	parent   context.Context
	iterable Iterable[T]
}

// NewOptionalSingleImpl create OptionalSingleImpl
func NewOptionalSingleImpl[T any](iterable Iterable[T]) OptionalSingleImpl[T] {
	// this is new functionality due to iterable not being exported
	//
	return OptionalSingleImpl[T]{
		iterable: iterable,
	}
}

// Get returns the item or rxgo.OptionalEmpty. The error returned is if the context has been cancelled.
// This method is blocking.
func (o *OptionalSingleImpl[T]) Get(opts ...Option[T]) (Item[T], error) {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	optionalSingleEmpty := Item[T]{}
	observe := o.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return optionalSingleEmpty, ctx.Err()
		case v, ok := <-observe:
			if !ok {
				return optionalSingleEmpty, nil
			}

			return v, nil
		}
	}
}

// Map transforms the items emitted by an OptionalSingle by applying a function to each item.
func (o *OptionalSingleImpl[T]) Map(apply Func[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &mapOperatorOptionalSingle[T]{apply: apply}
	}, forceSeq, bypassGather, opts...)
}

// Observe observes an OptionalSingle by returning its channel.
func (o *OptionalSingleImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return o.iterable.Observe(opts...)
}

type mapOperatorOptionalSingle[T any] struct {
	apply Func[T]
}

func (op *mapOperatorOptionalSingle[T]) next(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// TODO: no longer needed: if !item.IsNumeric() {
	// 	panic(fmt.Errorf("not a number (%v)", item))
	// }
	//
	res, err := op.apply(ctx, item.V)
	if err != nil {
		dst <- Error[T](err)

		operatorOptions.stop()

		return
	}

	dst <- Of(res)
}

func (op *mapOperatorOptionalSingle[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperatorOptionalSingle[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *mapOperatorOptionalSingle[T]) gatherNext(_ context.Context,
	item Item[T], dst chan<- Item[T], _ operatorOptions[T],
) {
	// TODO: --> switch item.V.(type) {
	// case *mapOperatorOptionalSingle[T]:
	// 	return
	// }
	dst <- item

	panic("mapOperatorOptionalSingle: please check the commented out switch statement")
}

// Run creates an observer without consuming the emitted items.
func (o *OptionalSingleImpl[T]) Run(opts ...Option[T]) Disposed {
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
