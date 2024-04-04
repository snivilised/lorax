package rx

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
	// no longer needed: if !item.IsNumeric() {
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
	// --> switch item.V.(type) {
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
