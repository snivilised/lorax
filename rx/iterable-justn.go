package rx

import (
	"context"
)

type justNIterable[T any] struct {
	items []int
	opts  []Option[T]
}

func newJustNIterable[T any](numbers ...int) func(opts ...Option[T]) Iterable[T] {
	// do we turn items back into ints?
	//
	return func(opts ...Option[T]) Iterable[T] {
		return &justNIterable[T]{
			items: numbers,
			opts:  opts,
		}
	}
}

func (i *justNIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()
	items := make([]int, 0, len(i.items))
	items = append(items, i.items...)

	go SendNumbers(option.buildContext(emptyContext), next, CloseChannel,
		items...,
	)

	return next
}

// SendItems is an utility function that send a list of items and indicate a
// strategy on whether to close the channel once the function completes.
func SendNumbers[T any](ctx context.Context,
	ch chan<- Item[T], strategy CloseChannelStrategy, numbers ...int,
) {
	if strategy == CloseChannel {
		defer close(ch)
	}

	for _, current := range numbers {
		Num[T](current).SendContext(ctx, ch)
	}
}
