package rx

import (
	"context"
	"sync"
)

type createIterable[T any] struct {
	next                   <-chan Item[T]
	opts                   []Option[T]
	subscribers            []chan Item[T]
	mutex                  sync.RWMutex
	producerAlreadyCreated bool
}

func newCreateIterable[T any](fs []Producer[T], opts ...Option[T]) Iterable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(emptyContext)

	go func() {
		defer close(next)

		for _, f := range fs {
			f(ctx, next)
		}
	}()

	return &createIterable[T]{
		opts: opts,
		next: next,
	}
}

func (i *createIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	mergedOptions := append(i.opts, opts...) //nolint:gocritic // ignore
	option := parseOptions(mergedOptions...)

	if !option.isConnectable() {
		return i.next
	}

	if option.isConnectOperation() {
		i.connect(option.buildContext(emptyContext))
		return nil
	}

	ch := option.buildChannel()

	i.mutex.Lock()
	i.subscribers = append(i.subscribers, ch)
	i.mutex.Unlock()

	return ch
}

func (i *createIterable[T]) connect(ctx context.Context) {
	i.mutex.Lock()
	if !i.producerAlreadyCreated {
		go i.produce(ctx)
		i.producerAlreadyCreated = true
	}
	i.mutex.Unlock()
}

func (i *createIterable[T]) produce(ctx context.Context) {
	defer func() {
		i.mutex.RLock()
		for _, subscriber := range i.subscribers {
			close(subscriber)
		}
		i.mutex.RUnlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-i.next:
			if !ok {
				return
			}

			i.mutex.RLock()
			for _, subscriber := range i.subscribers {
				subscriber <- item
			}
			i.mutex.RUnlock()
		}
	}
}
