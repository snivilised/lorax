// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ants

import (
	"context"
	"time"
)

type workerStack struct {
	items  []worker
	expiry []worker
}

func newWorkerStack(size int) *workerStack {
	return &workerStack{
		items: make([]worker, 0, size),
	}
}

func (wq *workerStack) len() int {
	return len(wq.items)
}

func (wq *workerStack) isEmpty() bool {
	return len(wq.items) == 0
}

func (wq *workerStack) insert(w worker) error {
	wq.items = append(wq.items, w)
	return nil
}

func (wq *workerStack) detach() worker {
	l := wq.len()
	if l == 0 {
		return nil
	}

	w := wq.items[l-1]
	wq.items[l-1] = nil // avoid memory leaks
	wq.items = wq.items[:l-1]

	return w
}

func (wq *workerStack) refresh(duration time.Duration) []worker {
	n := wq.len()
	if n == 0 {
		return nil
	}

	expiryTime := time.Now().Add(-duration)
	index := wq.search(0, n-1, expiryTime)

	wq.expiry = wq.expiry[:0]
	if index != -1 {
		wq.expiry = append(wq.expiry, wq.items[:index+1]...)
		m := copy(wq.items, wq.items[index+1:])
		for i := m; i < n; i++ {
			wq.items[i] = nil
		}
		wq.items = wq.items[:m]
	}
	return wq.expiry
}

func (wq *workerStack) search(l, r int, expiryTime time.Time) int {
	for l <= r {
		mid := l + ((r - l) >> 1) // avoid overflow when computing mid
		if expiryTime.Before(wq.items[mid].lastUsedTime()) {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return r
}

func (wq *workerStack) reset(ctx context.Context) {
	for i := 0; i < wq.len(); i++ {
		wq.items[i].finish(ctx)
		wq.items[i] = nil
	}
	wq.items = wq.items[:0]
}
