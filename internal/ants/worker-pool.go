package ants

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type workerPool struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// lock for protecting the worker queue.
	lock sync.Locker

	// workers is a slice that store the available workers.
	workers workerQueue

	// state is used to notice the pool to closed itself.
	state int32

	// cond for waiting to get an idle worker.
	cond *sync.Cond

	// workerCache speeds up the obtainment of a usable worker in function:retrieveWorker.
	workerCache sync.Pool

	// waiting is the number of the goroutines already been blocked on pool.Invoke(), protected by pool.lock
	waiting int32

	purgeDone int32
	stopPurge context.CancelFunc

	ticktockDone int32
	stopTicktock context.CancelFunc

	now atomic.Value

	o *Options
}

// Running returns the number of workers currently running.
func (p *workerPool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the number of available workers, -1 indicates this pool
// is unlimited.
func (p *workerPool) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}

	return c - p.Running()
}

// Waiting returns the number of tasks waiting to be executed.
func (p *workerPool) Waiting() int {
	return int(atomic.LoadInt32(&p.waiting))
}

// Cap returns the capacity of this pool.
func (p *workerPool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool, note that it is noneffective to
// the infinite or pre-allocation pool.
func (p *workerPool) Tune(size int) {
	capacity := p.Cap()
	if capacity == -1 || size <= 0 || size == capacity || p.o.PreAlloc {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	if size > capacity {
		if size-capacity == 1 {
			p.cond.Signal()

			return
		}
		p.cond.Broadcast()
	}
}

// IsClosed indicates whether the pool is closed.
func (p *workerPool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == CLOSED
}

// Release closes this pool and releases the worker queue.
func (p *workerPool) Release(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&p.state, OPENED, CLOSED) {
		return
	}

	if p.stopPurge != nil {
		p.stopPurge()
		p.stopPurge = nil
	}
	p.stopTicktock()
	p.stopTicktock = nil

	p.lock.Lock()
	p.workers.reset(ctx)
	p.lock.Unlock()
	// There might be some callers waiting in retrieveWorker(), so we need to
	// wake them up to prevent those callers blocking infinitely.
	p.cond.Broadcast()
}

// ReleaseTimeout is like Release but with a timeout, it waits all workers
// to exit before timing out.
func (p *workerPool) ReleaseTimeout(ctx context.Context, timeout time.Duration) error {
	purge := (!p.o.DisablePurge && p.stopPurge == nil)
	if p.IsClosed() || purge || p.stopTicktock == nil {
		return ErrPoolClosed
	}
	p.Release(ctx)

	endTime := time.Now().Add(timeout)
	for time.Now().Before(endTime) {
		if p.Running() == 0 &&
			(p.o.DisablePurge || atomic.LoadInt32(&p.purgeDone) == 1) &&
			atomic.LoadInt32(&p.ticktockDone) == 1 {
			return nil
		}
		time.Sleep(releaseTimeoutInterval * time.Millisecond)
	}

	return ErrTimeout
}

func (p *workerPool) addRunning(delta int) {
	atomic.AddInt32(&p.running, int32(delta))
}

func (p *workerPool) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

func (p *workerPool) GetOptions() *Options {
	return p.o
}
