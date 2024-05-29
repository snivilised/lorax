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
	"runtime/debug"
	"time"
)

// goWorker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type goWorker struct {
	// pool who owns this worker.
	pool *Pool

	// taskCh is a job should be done.
	taskCh TaskStream

	// lastUsed will be updated when putting a worker back into queue.
	lastUsed time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorker) run() {
	w.pool.addRunning(1)

	go func() {
		defer func() {
			w.pool.addRunning(-1)
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.o.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.o.Logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			w.pool.cond.Signal()
		}()

		for f := range w.taskCh {
			if f == nil { // ✨
				return
			}
			f()
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}

func (w *goWorker) finish(ctx context.Context) {
	select {
	case <-ctx.Done():
	case w.taskCh <- nil: // ✨:
	}
}

func (w *goWorker) lastUsedTime() time.Time {
	return w.lastUsed
}

func (w *goWorker) sendTask(ctx context.Context, fn TaskFunc) {
	select {
	case <-ctx.Done():
	case w.taskCh <- fn:
	}
}

func (w *goWorker) sendParam(context.Context, InputParam) {
	panic("unreachable")
}
