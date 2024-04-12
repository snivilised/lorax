package rx

type rangeIterable[T any] struct {
	start, count NumVal
	opts         []Option[T]
}

func newRangeIterable[T any](start, count NumVal, opts ...Option[T]) Iterable[T] {
	return &rangeIterable[T]{
		start: start,
		count: count,
		opts:  opts,
	}
}

func (i *rangeIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()

	go func() {
		for idx := i.start; idx <= i.start+i.count-1; idx++ {
			select {
			case <-ctx.Done():
				return
			case next <- Num[T](idx):
			}
		}
		close(next)
	}()

	return next
}
