package rx

// Amb takes several Observables, emit all of the items from only the first of these Observables
// to emit an item or notification.
func Amb[T any](observables []Observable[T], opts ...Option[T]) Observable[T] {
	_, _ = observables, opts

	panic("Amb: NOT-IMPL")
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

// Just creates an Observable with the provided items.
func Just[T any](values ...T) func(opts ...Option[T]) Observable[T] {
	return func(opts ...Option[T]) Observable[T] {
		return &ObservableImpl[T]{
			iterable: newJustIterable(values...)(opts...),
		}
	}
}

// JustSingle is like JustItem in that it is defined for a single item iterable
// but behaves like Just in that it returns a func.
// This is probably not required, just defined for experimental purposes for now.
func JustSingle[T any](value T, opts ...Option[T]) func(opts ...Option[T]) Single[T] {
	return func(_ ...Option[T]) Single[T] {
		return &SingleImpl[T]{
			iterable: newJustIterable(value)(opts...),
		}
	}
}

// JustItem creates a single from one item.
func JustItem[T any](value T, opts ...Option[T]) Single[T] {
	// Why does this not return a func, but Just does?
	//
	return &SingleImpl[T]{
		iterable: newJustIterable(value)(opts...),
	}
}

// Never creates an Observable that emits no items and does not terminate.
func Never[T any]() Observable[T] {
	next := make(chan Item[T])

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}
