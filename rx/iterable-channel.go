package rx

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
