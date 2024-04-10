package rx

type sliceIterable[T any] struct {
	items []Item[T]
	opts  []Option[T]
}

func newSliceIterable[T any](items []Item[T], opts ...Option[T]) Iterable[T] {
	return &sliceIterable[T]{
		items: items,
		opts:  opts,
	}
}

func (i *sliceIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()
	ctx := option.buildContext(emptyContext)

	go func() {
		for _, item := range i.items {
			select {
			case <-ctx.Done():
				return
			case next <- item:
			}
		}

		close(next)
	}()

	return next
}
