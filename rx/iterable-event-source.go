package rx

import (
	"context"
	"sync"
)

type eventSourceIterable[T any] struct {
	sync.RWMutex
	observers []chan Item[T]
	disposed  bool
	opts      []Option[T]
}

func newEventSourceIterable[T any](ctx context.Context,
	next <-chan Item[T], strategy BackPressureStrategy, opts ...Option[T],
) Iterable[T] {
	it := &eventSourceIterable[T]{
		observers: make([]chan Item[T], 0),
		opts:      opts,
	}

	go func() {
		defer func() {
			it.closeAllObservers()
		}()

		deliver := func(item Item[T]) (done bool) {
			it.RLock()
			defer it.RUnlock()

			switch strategy {
			default:
				fallthrough
			case Block:
				for _, observer := range it.observers {
					if !item.SendContext(ctx, observer) {
						return true
					}
				}
			case Drop:
				for _, observer := range it.observers {
					select {
					default:
					case <-ctx.Done():
						return true
					case observer <- item:
					}
				}
			}

			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-next:
				if !ok {
					return
				}

				if done := deliver(item); done {
					return
				}
			}
		}
	}()

	return it
}

func (i *eventSourceIterable[T]) closeAllObservers() {
	i.Lock()
	for _, observer := range i.observers {
		close(observer)
	}

	i.disposed = true
	i.Unlock()
}

func (i *eventSourceIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	next := option.buildChannel()

	i.Lock()
	if i.disposed {
		close(next)
	} else {
		i.observers = append(i.observers, next)
	}
	i.Unlock()

	return next
}
