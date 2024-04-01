package rx

type deferIterable[T any] struct {
	fs   []Producer[T]
	opts []Option[T]
}

func newDeferIterable[T any](f []Producer[T], opts ...Option[T]) Iterable[T] {
	return &deferIterable[T]{
		fs:   f,
		opts: opts,
	}
}

func (i *deferIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()
	ctx := option.buildContext(emptyContext)

	go func() {
		defer close(next)

		for _, f := range i.fs {
			f(ctx, next)
		}
	}()

	return next
}
