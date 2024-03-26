package rx

// Amb takes several Observables, emit all of the items from only the first of these Observables
// to emit an item or notification.
func Amb[I any](observables []Observable[I], opts ...Option[I]) Observable[I] {
	_, _ = observables, opts

	panic("Amb: NOT-IMPL")
}

// Empty creates an Observable with no item and terminate immediately.
func Empty[I any]() Observable[I] {
	next := make(chan Item[I])
	close(next)

	return &ObservableImpl[I]{
		iterable: newChannelIterable(next),
	}
}

// FromChannel creates a cold observable from a channel.
func FromChannel[I any](next <-chan Item[I], opts ...Option[I]) Observable[I] {
	option := parseOptions(opts...)
	ctx := option.buildContext(emptyContext)

	return &ObservableImpl[I]{
		parent:   ctx,
		iterable: newChannelIterable(next, opts...),
	}
}

// Just creates an Observable with the provided items.
func Just[I any](values ...I) func(opts ...Option[I]) Observable[I] {
	return func(opts ...Option[I]) Observable[I] {
		return &ObservableImpl[I]{
			iterable: newJustIterable(values...)(opts...),
		}
	}
}

// JustSingle is like JustItem in that it is defined for a single item iterable
// but behaves like Just in that it returns a func.
// This is probably not required, just defined for experimental purposes for now.
func JustSingle[I any](value I, opts ...Option[I]) func(opts ...Option[I]) Single[I] {
	return func(_ ...Option[I]) Single[I] {
		return &SingleImpl[I]{
			iterable: newJustIterable(value)(opts...),
		}
	}
}

// JustItem creates a single from one item.
func JustItem[I any](value I, opts ...Option[I]) Single[I] {
	// Why does this not return a func, but Just does?
	//
	return &SingleImpl[I]{
		iterable: newJustIterable(value)(opts...),
	}
}

// Never creates an Observable that emits no items and does not terminate.
func Never[I any]() Observable[I] {
	next := make(chan Item[I])

	return &ObservableImpl[I]{
		iterable: newChannelIterable(next),
	}
}
