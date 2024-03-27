package rx

type justIterable[T any] struct {
	items []any
	opts  []Option[T]
}

func newJustIterable[T any](items ...any) func(opts ...Option[T]) Iterable[T] {
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
	items := make([]any, 0, len(i.items))
	items = append(items, i.items...)

	go SendItems(option.buildContext(emptyContext), next, CloseChannel,
		items...,
	)

	return next
}
