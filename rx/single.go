package rx

import "context"

// Single is a observable with a single element.
type Single[I any] interface {
	Iterable[I]
	Filter(apply Predicate[I], opts ...Option[I]) OptionalSingle[I]
	Get(opts ...Option[I]) (Item[I], error)
	Map(apply Func[I], opts ...Option[I]) Single[I]
	Run(opts ...Option[I]) Disposed
}

// SingleImpl implements Single.
type SingleImpl[I any] struct {
	parent   context.Context
	iterable Iterable[I]
}

// Filter emits only those items from an Observable that pass a predicate test.
func (s *SingleImpl[I]) Filter(apply Predicate[I], opts ...Option[I]) OptionalSingle[I] {
	return optionalSingle(s.parent, s, func() operator[I] {
		return &filterOperatorSingle[I]{apply: apply}
	}, true, true, opts...)
}

// Get returns the item. The error returned is if the context has been cancelled.
// This method is blocking.
func (s *SingleImpl[I]) Get(opts ...Option[I]) (Item[I], error) {
	option := parseOptions(opts...)
	ctx := option.buildContext(s.parent)

	observe := s.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return Item[I]{}, ctx.Err()
		case v := <-observe:
			return v, nil
		}
	}
}

// Map transforms the items emitted by a Single by applying a function to each item.
func (s *SingleImpl[I]) Map(apply Func[I], opts ...Option[I]) Single[I] {
	return single(s.parent, s, func() operator[I] {
		return &mapOperatorSingle[I]{apply: apply}
	}, false, true, opts...)
}

type mapOperatorSingle[I any] struct {
	apply Func[I]
}

func (op *mapOperatorSingle[I]) next(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	res, err := op.apply(ctx, item.V)

	if err != nil {
		Error[I](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	Of(res).SendContext(ctx, dst)
}

func (op *mapOperatorSingle[I]) err(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperatorSingle[I]) end(_ context.Context, _ chan<- Item[I]) {
}

func (op *mapOperatorSingle[I]) gatherNext(ctx context.Context,
	item Item[I], dst chan<- Item[I], _ operatorOptions[I],
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
func (s *SingleImpl[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	return s.iterable.Observe(opts...)
}

type filterOperatorSingle[I any] struct {
	apply Predicate[I]
}

func (op *filterOperatorSingle[I]) next(ctx context.Context,
	item Item[I], dst chan<- Item[I], _ operatorOptions[I],
) {
	if op.apply(item.V) {
		item.SendContext(ctx, dst)
	}
}

func (op *filterOperatorSingle[I]) err(ctx context.Context,
	item Item[I], dst chan<- Item[I], operatorOptions operatorOptions[I],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *filterOperatorSingle[I]) end(_ context.Context, _ chan<- Item[I]) {
}

func (op *filterOperatorSingle[I]) gatherNext(_ context.Context,
	_ Item[I], _ chan<- Item[I], _ operatorOptions[I],
) {
}

// Run creates an observer without consuming the emitted items.
func (s *SingleImpl[I]) Run(opts ...Option[I]) Disposed {
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
