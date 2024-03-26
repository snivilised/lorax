package rx

type justIterable[T any] struct {
	items []T
	opts  []Option[T]
}

func newJustIterable[T any](items ...T) func(opts ...Option[T]) Iterable[T] {
	return func(opts ...Option[T]) Iterable[T] {
		return &justIterable[T]{
			items: items,
			opts:  opts,
		}
	}
}

func (i *justIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()
	items := make([]Item[T], 0, len(i.items))

	for _, item := range i.items {
		items = append(items, Of(item))
	}

	go SendItems(option.buildContext(emptyContext), next, CloseChannel, items...)

	return next
}
