package rx

import (
	"context"
	"reflect"
)

func isZero[T any](limit T) bool {
	val := reflect.ValueOf(limit).Interface()
	zero := reflect.Zero(reflect.TypeOf(limit)).Interface()

	return val != zero
}

func (o *ObservableImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return o.iterable.Observe(opts...)
}

// Connect instructs a connectable Observable to begin emitting items to its subscribers.
func (o *ObservableImpl[T]) Connect(ctx context.Context) (context.Context, Disposable) {
	ctx, cancel := context.WithCancel(ctx)
	o.Observe(WithContext[T](ctx), connect[T]())

	return ctx, Disposable(cancel)
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

// Max determines and emits the maximum-valued item emitted by an Observable according to a comparator.
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
	max        T
}

func (op *maxOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	if op.comparator(op.max, item.V) < 0 {
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
	// no longer needed: if !item.IsNumeric() {
	// 	panic(fmt.Errorf("not a number (%v)", item))
	// }
	//
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
	// switch item.V.(type) {
	// case *mapOperator:
	// 	return
	// }
	// TODO: check above switch not required
	item.SendContext(ctx, dst)
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

	if op.comparator(op.min, item.V) > 0 {
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
