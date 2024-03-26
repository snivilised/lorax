package rx

import (
	"context"
)

// OptionalSingle is an optional single.
type OptionalSingle[I any] interface {
	Iterable[I]
	Get(opts ...Option[I]) (Item[I], error)
	Map(apply Func[I], opts ...Option[I]) OptionalSingle[I]
	Run(opts ...Option[I]) Disposed
}

// OptionalSingleImpl implements OptionalSingle.
type OptionalSingleImpl[I any] struct {
	parent   context.Context
	iterable Iterable[I]
}

// NewOptionalSingleImpl create OptionalSingleImpl
func NewOptionalSingleImpl[I any](iterable Iterable[I]) OptionalSingleImpl[I] {
	// this is new functionality due to iterable not being exported
	//
	return OptionalSingleImpl[I]{
		iterable: iterable,
	}
}

// Get returns the item or rxgo.OptionalEmpty. The error returned is if the context has been cancelled.
// This method is blocking.
func (o *OptionalSingleImpl[I]) Get(opts ...Option[I]) (Item[I], error) {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	optionalSingleEmpty := Item[I]{}
	observe := o.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return Item[I]{}, ctx.Err()
		case v, ok := <-observe:
			if !ok {
				return optionalSingleEmpty, nil
			}

			return v, nil
		}
	}
}

// Map transforms the items emitted by an OptionalSingle by applying a function to each item.
func (o *OptionalSingleImpl[I]) Map(apply Func[I], opts ...Option[I]) OptionalSingle[I] {
	return optionalSingle(o.parent, o, func() operator[I] {
		return &mapOperatorOptionalSingle[I]{apply: apply}
	}, false, true, opts...)
}

// Observe observes an OptionalSingle by returning its channel.
func (o *OptionalSingleImpl[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	return o.iterable.Observe(opts...)
}

type mapOperatorOptionalSingle[I any] struct {
	apply Func[I]
}

func (op *mapOperatorOptionalSingle[I]) next(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	res, err := op.apply(ctx, item.V)
	if err != nil {
		dst <- Error[I](err)

		operatorOptions.stop()

		return
	}
	dst <- Of(res)
}

func (op *mapOperatorOptionalSingle[I]) err(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperatorOptionalSingle[I]) end(_ context.Context, _ chan<- Item[I]) {
}

func (op *mapOperatorOptionalSingle[I]) gatherNext(_ context.Context,
	item Item[I], dst chan<- Item[I], _ operatorOptions[I],
) {
	// --> switch item.V.(type) {
	// case *mapOperatorOptionalSingle[I]:
	// 	return
	// }
	dst <- item

	panic("mapOperatorOptionalSingle: please check the commented out switch statement")
}

// Run creates an observer without consuming the emitted items.
func (o *OptionalSingleImpl[I]) Run(opts ...Option[I]) Disposed {
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
