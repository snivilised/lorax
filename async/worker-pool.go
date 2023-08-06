package async

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/google/uuid"
)

/*
ref: https://levelup.gitconnected.com/how-to-use-context-to-manage-your-goroutines-like-a-boss-ef1e478919e6

func main() {
    // Create a new context.
    parent, cancelParent := context.WithCancel(context.Background())
    // Derive child contexts from parent.
    childA, _ := context.WithTimeout(parent, 5 * time.Second)
    childB, _ := context.WithDeadline(parent, time.Now().Add(1 * time.Minute)
    go func() {
        <-childA.Done()
        <-childB.Done()
        fmt.Println("All children are done")
    }()
    // Cancel parent make all children are cancelled.
    cancelParent()
}
// -> Result: All children are done

* context.WithCancel(parentContext) creates a new context which completes when
	the returned cancel function is called or when the parent's context finishes,
	whichever happens first.

* context.WithTimeout(contextContext, 5 * time.Second) creates a new context
	which finishes when the returned cancel function is called or when it exceeds
	timeout or when the parent's context finishes, whichever happens first.

* context.WithDeadline(parentContext, time.Now().Add(1 * time.Minute) creates a
	new context which finishes when the returned cancel function deadline expires
	or when the parent's context completes, whichever happens first.

See also: https://pkg.go.dev/context#example_WithCancel
				: https://go.dev/blog/pprof
				: https://levelup.gitconnected.com/how-to-use-context-to-manage-your-goroutines-like-a-boss-ef1e478919e6
				: https://blog.logrocket.com/functional-programming-in-go/

*/

/*
Channels in play:
	- jobs (input)
	- results (output)
	- errors (output)
	- cancel (signal)
	- done (signals no more new work)

The effect we want to create is similar to the design of io_uring
in linux

We want the main thread to perform a close on the jobs channel when
there no more work required. This closure should not interrupt the
execution of the existing workload.

The question is, do we want to use a new GR to send jobs to the pool
or do we want this to be blocking for the main GR?

1) If we have a dedicated dispatcher GR, then that implies the main thread
could freely submit all jobs it can find without being throttled. The downside
of this is that we could have a large build up of outstanding jobs resulting
in higher memory consumption. We would need a way to wait for all jobs to be
completed, ie when there are no more workers. This could be achieved with
a wait group. The problem with a wait group is that we could accidentally
reach 0 in the wait group even though there are still jobs outstanding. This is
a race condition which would arise because a job in the queue is not taken up
before all existing workers have exited. We could alleviate this by adding
an extra entry into the wait group but then how do you get down to 0?

2) Main GR sends jobs into a buffered channel and blocks when full. This seems
like the more sensible option. The main GR would be throttled by the number of
active workers and the job queue would not grow excessively consuming
memory judiciously.

However, if we have a results channel that must be read from, then we can't
have the main GR limited by the size of the worker pool, because if we do, we'll
still suffer frm the problem of memory build up, but this would be a build up
on the output, ie of the results channel.
GR(main) --> jobsCh: this is blocking after channel is full

3)

ProducerGR(observable):
	- writes to job channel

PoolGR(workers):
	- reads from job channel

ConsumerGR(observer):
	- reads from results channel
	- reads from errors channel

Both the Producer and the Consumer should be started up immediately as
separate GRs, distinct from the main GR.

	* ProducerGR(observable) --> owns the job channel and should be free to close it
	when no more work is available.

	* PoolGR(workers) --> the pool owns the output channels

So, the next question is, how does the pool know when to close the output channels?
In theory, this should be when the jobs queue is empty and the current pool of
workers is empty. This realisation now makes us discover what the worker is. The
worker is effectively a handle to the go routine which is stored in a scoped collection.
Operations on this collection should be done via a channel, where we send a pointer
to the collection thru the channel. This collection should probably be a map, whose key
is a uniquely generated ID (see "github.com/google/uuid"). When the map is empty, we
know there are no more workers active to send to the outputs, therefore we can close them.

---
- To signal an event without sending data, we can use sync.Cond. (See
page 53. ). We could use Cond to signal no more work and no more results.

- The results channel must be optional, because a client may define work in which a result
is of no value. In this case, the pool must decide how to define closure. Perhaps it creates
a dummy consumer.
*/

// The WorkerPool owns the resultOut channel, because it is the only entity that knows
// when all workers have completed their work due to the finished channel, which it also
// owns.

type WorkerPool[I, R any] struct {
	fn         Executive[I, R]
	noWorkers  int
	JobsCh     JobStream[I]
	ResultsCh  ResultStream[R]
	CancelCh   CancelStream
	Quit       *sync.WaitGroup
	pool       workersCollection[I, R]
	finishedCh FinishedStream
}

type NewWorkerPoolParams[I, R any] struct {
	Exec   Executive[I, R]
	JobsCh JobStream[I]
	Cancel CancelStream
	Quit   *sync.WaitGroup
}

func NewWorkerPool[I, R any](params *NewWorkerPoolParams[I, R]) *WorkerPool[I, R] {
	wp := &WorkerPool[I, R]{
		fn:        params.Exec,
		noWorkers: runtime.NumCPU(),
		JobsCh:    params.JobsCh,
		CancelCh:  params.Cancel,
		Quit:      params.Quit,

		// workers collection might not be necessary; only using here at the
		// moment, so it is easy to track how many workers are running at
		// any 1 time.
		//
		pool:       make(workersCollection[I, R]),
		finishedCh: make(FinishedStream, DefaultChSize),
	}

	return wp
}

// Run
func (p *WorkerPool[I, R]) Run(ctx context.Context, resultsOut ResultStreamOut[R]) {
	defer func() {
		fmt.Printf("<--- WorkerPool finished (Quit). 🧊🧊🧊\n")
		p.Quit.Done()
		close(resultsOut)
	}()
	fmt.Println("---> 🧊 WorkerPool.Run")

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Println("---> 🧊 WorkerPool.Run - done received ☢️☢️☢️")

			running = false

		case job, ok := <-p.JobsCh:
			if ok {
				fmt.Println("---> 🧊 WorkerPool.Run - new job received")

				p.dispatch(ctx, &workerInfo[I, R]{
					job:         job,
					resultsOut:  resultsOut,
					finishedOut: p.finishedCh,
				})
			} else {
				running = false
			}

		case workerID := <-p.finishedCh:
			fmt.Printf("---> 🧊 WorkerPool.Run - worker(%v) finished\n", workerID)
			delete(p.pool, workerID)
		}
	}

	// we still need to wait for all workers to finish ...
	//
	p.drain(ctx)
}

func (p *WorkerPool[I, R]) drain(ctx context.Context) {
	// The remaining number of workers displayed here is not necessarily
	// accurate.
	//
	fmt.Printf(
		"!!!! 🧊 WorkerPool.drain - waiting for remaining workers: %v (#GRs: %v); 🧊🧊🧊 \n",
		len(p.pool), runtime.NumGoroutine(),
	)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false

		case workerID := <-p.finishedCh:
			fmt.Printf("---> 🧊 WorkerPool.drain - worker(%v) finished\n", workerID)
			delete(p.pool, workerID)

			if len(p.pool) == 0 {
				running = false
			}
		}
	}
}

func (p *WorkerPool[I, R]) dispatch(ctx context.Context, info *workerInfo[I, R]) {
	w := &worker[I, R]{
		id: WorkerID("WORKER-ID:" + uuid.NewString()),
		fn: p.fn,
	}
	p.pool[w.id] = w
	fmt.Printf("---> 🧊 (pool-size: %v) dispatch worker: id-'%v'\n", len(p.pool), w.id)

	go w.accept(ctx, info) // BREAKS: when cancellation occurs, send on closed chan
}
