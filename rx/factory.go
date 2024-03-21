package rx

// Amb takes several Observables, emit all of the items from only the first of these Observables
// to emit an item or notification.
func Amb[I any](observables []Observable[I], opts ...Option[I]) Observable[I] {
	_, _ = observables, opts

	panic("Amb: NOT-IMPL")
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

func parseOptions[I any](opts ...Option[I]) Option[I] {
	o := new(funcOption[I])
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}
