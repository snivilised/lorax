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
