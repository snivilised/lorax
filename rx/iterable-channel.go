package rx

import (
	"context"
	"sync"
)

type channelIterable[T any] struct {
	next                   <-chan Item[T]
	opts                   []Option[T]
	subscribers            []chan Item[T]
	mutex                  sync.RWMutex
	producerAlreadyCreated bool
}

func newChannelIterable[T any](next <-chan Item[T], opts ...Option[T]) Iterable[T] {
	return &channelIterable[T]{
		next:        next,
		subscribers: make([]chan Item[T], 0),
		opts:        opts,
	}
}

func (i *channelIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	mergedOptions := make([]Option[T], 0, len(opts))
	copy(mergedOptions, opts)
	mergedOptions = append(mergedOptions, opts...)

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

func (i *channelIterable[T]) connect(ctx context.Context) {
	i.mutex.Lock()
	if !i.producerAlreadyCreated {
		go i.produce(ctx)
		i.producerAlreadyCreated = true
	}
	i.mutex.Unlock()
}

func (i *channelIterable[T]) produce(ctx context.Context) {
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
