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
