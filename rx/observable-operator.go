package rx

import (
	"context"
	"reflect"
)

func (o *ObservableImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return o.iterable.Observe(opts...)
}

// Max determines and emits the maximum-valued item emitted by an Observable according to a comparator.
func (o *ObservableImpl[T]) Max(comparator Comparator[T],
	opts ...Option[T],
) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &maxOperator[T]{
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

func isLimitDefined[T any](limit T) bool {
	val := reflect.ValueOf(limit).Interface()
	zero := reflect.Zero(reflect.TypeOf(limit)).Interface()

	return val != zero
}

type maxOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	max        T
}

func (op *maxOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	// TODO(check): if op.max == nil {
	// 	op.max = item.V
	// } else {
	// 	if op.comparator(op.max, item.V) < 0 {
	// 		op.max = item.V
	// 	}
	// }
	if !isLimitDefined(op.max) || (op.comparator(op.max, item.V) < 0) {
		op.max = item.V
	}
}

func (op *maxOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *maxOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		Of(op.max).SendContext(ctx, dst)
	}
}

func (op *maxOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*maxOperator).max), dst, operatorOptions)÷
	op.next(ctx, Of(item.V), dst, operatorOptions)
}

// Min determines and emits the minimum-valued item emitted by an Observable
// according to a comparator.
func (o *ObservableImpl[T]) Min(comparator Comparator[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &minOperator[T]{
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

type minOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	min        T
	limit      func(value T) bool // represents min or max
}

func (op *minOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	// TODO(check): if op.min == nil {
	// 	op.min = item.V
	// } else {
	// 	if op.comparator(op.min, item.V) > 0 {
	// 		op.min = item.V
	// 	}
	// }
	if !isLimitDefined(op.min) || (op.comparator(op.min, item.V) > 0) {
		op.min = item.V
	}
}

func (op *minOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *minOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		Of(op.min).SendContext(ctx, dst)
	}
}

func (op *minOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*minOperator).min), dst, operatorOptions)
	op.next(ctx, Of(item.V), dst, operatorOptions)
}
