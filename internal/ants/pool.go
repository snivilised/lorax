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
	"sync"
	"sync/atomic"
	"time"

	"github.com/snivilised/lorax/internal/ants/async"
)

// Pool accepts the tasks and process them concurrently,
// it limits the total of goroutines to a given number by recycling goroutines.
type Pool struct {
	workerPool
}

// purgeStaleWorkers clears stale workers periodically, it runs in an
// individual goroutine, as a scavenger.
func (p *Pool) purgeStaleWorkers(purgeCtx context.Context) {
	ticker := time.NewTicker(p.o.ExpiryDuration)

	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.purgeDone, 1)
	}()

	for {
		select {
		case <-purgeCtx.Done():
			return
		case <-ticker.C:
		}

		if p.IsClosed() {
			break
		}

		var isDormant bool
		p.lock.Lock()
		staleWorkers := p.workers.refresh(p.o.ExpiryDuration)
		n := p.Running()
		isDormant = n == 0 || n == len(staleWorkers)
		p.lock.Unlock()

		// Notify obsolete workers to stop.
		// This notification must be outside the p.lock, since w.task
		// may be blocking and may consume a lot of time if many workers
		// are located on non-local CPUs.
		for i := range staleWorkers {
			staleWorkers[i].finish(purgeCtx)
			staleWorkers[i] = nil
		}

		// There might be a situation where all workers have been cleaned
		// up (no worker is running), while some invokers still are stuck
		// in p.cond.Wait(), then we need to awake those invokers.
		if isDormant && p.Waiting() > 0 {
			p.cond.Broadcast()
		}
	}
}

// ticktock is a goroutine that updates the current time in the pool regularly.
func (p *Pool) ticktock(ticktockCtx context.Context) {
	ticker := time.NewTicker(nowTimeUpdateInterval)
	defer func() {
		ticker.Stop()
		atomic.StoreInt32(&p.ticktockDone, 1)
	}()

	for {
		select {
		case <-ticktockCtx.Done():
			return
		case <-ticker.C:
		}

		if p.IsClosed() {
			break
		}

		p.now.Store(time.Now())
	}
}

func (p *Pool) goPurge(ctx context.Context) {
	if p.o.DisablePurge {
		return
	}

	// Start a goroutine to clean up expired workers periodically.
	var purgeCtx context.Context
	purgeCtx, p.stopPurge = context.WithCancel(ctx)
	go p.purgeStaleWorkers(purgeCtx)
}

func (p *Pool) goTicktock(ctx context.Context) {
	p.now.Store(time.Now())
	var ticktockCtx context.Context
	ticktockCtx, p.stopTicktock = context.WithCancel(ctx)
	go p.ticktock(ticktockCtx)
}

func (p *Pool) nowTime() time.Time {
	return p.now.Load().(time.Time)
}

// NewPool instantiates a Pool with customized options.
func NewPool(ctx context.Context, size int, options ...Option) (*Pool, error) {
	if size <= 0 {
		size = -1
	}

	opts := loadOptions(options...)

	if !opts.DisablePurge {
		if expiry := opts.ExpiryDuration; expiry < 0 {
			return nil, ErrInvalidPoolExpiry
		} else if expiry == 0 {
			opts.ExpiryDuration = DefaultCleanIntervalTime
		}
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &Pool{
		workerPool: workerPool{
			capacity: int32(size),
			lock:     async.NewSpinLock(),
			o:        opts,
		},
	}

	p.workerCache.New = func() interface{} { // interface{} => sync.Pool api
		return &goWorker{
			pool:   p,
			taskCh: make(chan TaskFunc, workerChanCap),
		}
	}

	if p.o.PreAlloc {
		if size == -1 {
			return nil, ErrInvalidPreAllocSize
		}
		p.workers = newWorkerQueue(queueTypeLoopQueue, size)
	} else {
		p.workers = newWorkerQueue(queueTypeStack, 0)
	}

	p.cond = sync.NewCond(p.lock)

	p.goPurge(ctx)
	p.goTicktock(ctx)

	return p, nil
}

// Submit submits a task to this pool.
//
// Note that you are allowed to call Pool.Submit() from the current
// Pool.Submit(), but what calls for special attention is that you will
// get blocked with the last Pool.Submit() call once the current Pool
// runs out of its capacity, and to avoid this, you should instantiate
// a Pool with ants.WithNonblocking(true).
func (p *Pool) Submit(ctx context.Context, task TaskFunc) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}

	w, err := p.retrieveWorker()
	if w != nil {
		w.sendTask(ctx, task)
	}

	return err
}

// Reboot reboots a closed pool.
func (p *Pool) Reboot(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&p.state, CLOSED, OPENED) {
		atomic.StoreInt32(&p.purgeDone, 0)
		p.goPurge(ctx)
		atomic.StoreInt32(&p.ticktockDone, 0)
		p.goTicktock(ctx)
	}
}

// retrieveWorker returns an available worker to run the tasks.
func (p *Pool) retrieveWorker() (w worker, err error) {
	p.lock.Lock() // why isn't the unlock just deferred?

retry:
	// First try to fetch the worker from the queue.
	if w = p.workers.detach(); w != nil {
		p.lock.Unlock()

		return //nolint:nakedret // wtf
	}

	// If the worker queue is empty, and we don't run out of the pool capacity,
	// then just spawn a new worker goroutine.
	if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		p.lock.Unlock()
		w, _ = p.workerCache.Get().(*goWorker)
		w.run()

		return //nolint:nakedret // wtf
	}

	// Bail out early if it's in nonblocking mode or the number of pending
	// callers reaches the maximum limit value.
	full := p.Waiting() >= p.o.MaxBlockingTasks
	exceeded := p.o.MaxBlockingTasks != 0 && full

	if p.o.Nonblocking || exceeded {
		p.lock.Unlock()

		return nil, ErrPoolOverload
	}

	// Otherwise, we'll have to keep them blocked and wait for at least one
	// worker to be put back into pool.
	p.addWaiting(1)
	p.cond.Wait() // block and wait for an available worker
	p.addWaiting(-1)

	if p.IsClosed() {
		p.lock.Unlock()

		return nil, ErrPoolClosed
	}

	goto retry
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *goWorker) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		p.cond.Broadcast()

		return false
	}

	worker.lastUsed = p.nowTime()

	p.lock.Lock()
	// To avoid memory leaks, add a double check in the lock scope.
	// Issue: https://github.com/panjf2000/ants/issues/113
	if p.IsClosed() {
		p.lock.Unlock()

		return false
	}

	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()

		return false
	}
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available
	// worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()

	return true
}
