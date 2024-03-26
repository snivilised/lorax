package rx

type justIterable[I any] struct {
	items []I
	opts  []Option[I]
}

func newJustIterable[I any](items ...I) func(opts ...Option[I]) Iterable[I] {
	return func(opts ...Option[I]) Iterable[I] {
		return &justIterable[I]{
			items: items,
			opts:  opts,
		}
	}
}

func (i *justIterable[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()
	items := make([]Item[I], 0, len(i.items))

	for _, item := range i.items {
		items = append(items, Of(item))
	}

	go SendItems(option.buildContext(emptyContext), next, CloseChannel, items...)

	return next
}
