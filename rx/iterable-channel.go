package rx

import (
	"context"
	"sync"
)

type channelIterable[I any] struct {
	next                   <-chan Item[I]
	opts                   []Option[I]
	subscribers            []chan Item[I]
	mutex                  sync.RWMutex
	producerAlreadyCreated bool
}

func newChannelIterable[I any](next <-chan Item[I], opts ...Option[I]) Iterable[I] {
	return &channelIterable[I]{
		next:        next,
		subscribers: make([]chan Item[I], 0),
		opts:        opts,
	}
}

func (i *channelIterable[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	mergedOptions := make([]Option[I], 0, len(opts))
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

func (i *channelIterable[I]) connect(ctx context.Context) {
	i.mutex.Lock()
	if !i.producerAlreadyCreated {
		go i.produce(ctx)
		i.producerAlreadyCreated = true
	}
	i.mutex.Unlock()
}

func (i *channelIterable[I]) produce(ctx context.Context) {
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
