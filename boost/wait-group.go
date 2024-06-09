package boost

import (
	"sync"
	"sync/atomic"
)

// Tracker
type Tracker func(count int32)

// TrackWaitGroup returns a trackable wait group for the native sync
// wait group specified; useful for debugging purposes.
func TrackWaitGroup(wg *sync.WaitGroup, add, done Tracker) WaitGroup {
	return &TrackableWaitGroup{
		wg:   wg,
		add:  add,
		done: done,
	}
}

// TrackableWaitGroup
type TrackableWaitGroup struct {
	wg        *sync.WaitGroup
	counter   int32
	add, done Tracker
}

// Add
func (t *TrackableWaitGroup) Add(delta int) {
	n := atomic.AddInt32(&t.counter, int32(delta))
	t.wg.Add(delta)
	t.add(n)
}

// Done
func (t *TrackableWaitGroup) Done() {
	n := atomic.AddInt32(&t.counter, int32(-1))
	t.wg.Done()
	t.done(n)
}

// Wait
func (t *TrackableWaitGroup) Wait() {
	t.wg.Wait()
}

func (t *TrackableWaitGroup) Count() int32 {
	return atomic.LoadInt32(&t.counter)
}
