package rx

import (
	"context"
	"reflect"
)

func (o *ObservableImpl[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	return o.iterable.Observe(opts...)
}

// Max determines and emits the maximum-valued item emitted by an Observable according to a comparator.
func (o *ObservableImpl[I]) Max(comparator Comparator[I],
	opts ...Option[I],
) OptionalSingle[I] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[I] {
		return &maxOperator[I]{
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

func isLimitDefined[I any](limit I) bool {
	val := reflect.ValueOf(limit).Interface()
	zero := reflect.Zero(reflect.TypeOf(limit)).Interface()

	return val != zero
}

type maxOperator[I any] struct {
	comparator Comparator[I]
	empty      bool
	max        I
}

func (op *maxOperator[I]) next(_ context.Context,
	item Item[I], _ chan<- Item[I], _ operatorOptions[I],
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

func (op *maxOperator[I]) err(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *maxOperator[I]) end(ctx context.Context, dst chan<- Item[I]) {
	if !op.empty {
		Of(op.max).SendContext(ctx, dst)
	}
}

func (op *maxOperator[I]) gatherNext(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*maxOperator).max), dst, operatorOptions)÷
	op.next(ctx, Of(item.V), dst, operatorOptions)
}

// Min determines and emits the minimum-valued item emitted by an Observable
// according to a comparator.
func (o *ObservableImpl[I]) Min(comparator Comparator[I], opts ...Option[I]) OptionalSingle[I] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[I] {
		return &minOperator[I]{
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

type minOperator[I any] struct {
	comparator Comparator[I]
	empty      bool
	min        I
	limit      func(value I) bool // represents min or max
}

func (op *minOperator[I]) next(_ context.Context,
	item Item[I], _ chan<- Item[I], _ operatorOptions[I],
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

func (op *minOperator[I]) err(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *minOperator[I]) end(ctx context.Context, dst chan<- Item[I]) {
	if !op.empty {
		Of(op.min).SendContext(ctx, dst)
	}
}

func (op *minOperator[I]) gatherNext(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*minOperator).min), dst, operatorOptions)
	op.next(ctx, Of(item.V), dst, operatorOptions)
}
