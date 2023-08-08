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

// privateWpInfo contains any state that needs to be mutated in a non concurrent manner
// and therefore should be exclusively accessed by a single go routine. Actually, due to
// our ability to compose functionality with channels as opposed to shared state, the
// pool does not contain any state that is accessed directly or indirectly from other
// go routines. But in the case of the actual core pool, it is mutated without synchronisation
// and hence should only ever be accessed by the worker pool GR in contrast to all the
// other members of WorkerPool. This is an experimental pattern, the purpose of which
// is the clearly indicate what state can be accessed in different concurrency contexts,
// to ensure future updates can be applied with minimal cognitive overload.
//
// There is another purpose for privateWpInfo and that is to do with "confinement" as
// described on page 86 of CiG. The aim here is to use "lexical confinement" for
// duplex channel definitions, so although a channel is thread safe so ordinarily
// would not be a candidate member of privateWpInfo, a duplex channel ought to be
// protected from accidentally being used incorrectly, ie trying to write to a channel
// that is meant to be read only. So methods that use a channel should now receive the
// channel through a method parameter (defined as either chan<-, or <-chan), rather
// than be expected to simply access the member variable directly. This clearly signals
// that any channel defined in privateWpInfo should never to accessed directly (other
// than for passing it to another method). This is an experimental convention that
// I'm establishing for all snivilised projects.
type privateWpInfo[I, R any] struct {
	pool          workersCollection[I, R]
	workersJobsCh chan Job[I]
	finishedCh    FinishedStream
	cancelCh      CancelStream
}

// WorkerPool owns the resultOut channel, because it is the only entity that knows
// when all workers have completed their work due to the finished channel, which it also
// owns.
type WorkerPool[I, R any] struct {
	private        privateWpInfo[I, R]
	fn             Executive[I, R]
	noWorkers      int
	SourceJobsChIn <-chan Job[I]

	Quit *sync.WaitGroup
}

type NewWorkerPoolParams[I, R any] struct {
	NoWorkers int
	Exec      Executive[I, R]
	JobsCh    chan Job[I]
	CancelCh  CancelStream
	Quit      *sync.WaitGroup
}

func NewWorkerPool[I, R any](params *NewWorkerPoolParams[I, R]) *WorkerPool[I, R] {
	noWorkers := runtime.NumCPU()
	if params.NoWorkers > 1 && params.NoWorkers <= MaxWorkers {
		noWorkers = params.NoWorkers
	}

	wp := &WorkerPool[I, R]{
		private: privateWpInfo[I, R]{
			pool:          make(workersCollection[I, R], noWorkers),
			workersJobsCh: make(chan Job[I], noWorkers),
			finishedCh:    make(FinishedStream, noWorkers),
			cancelCh:      params.CancelCh,
		},
		fn:             params.Exec,
		noWorkers:      noWorkers,
		SourceJobsChIn: params.JobsCh,

		Quit: params.Quit,
	}

	return wp
}

// This helps to visualise the activity of the different work threads. Its easier to
// eyeball emojis than worker IDs.
var eyeballs = []string{
	"❤️", "💙", "💚", "💜", "💛", "🤍", "💖", "💗", "💝",
}

func (p *WorkerPool[I, R]) composeID() WorkerID {
	n := len(p.private.pool) + 1
	emoji := eyeballs[(n-1)%p.noWorkers]

	return WorkerID(fmt.Sprintf("(%v)WORKER-ID-%v:%v", emoji, n, uuid.NewString()))
}

func (p *WorkerPool[I, R]) Start(
	ctx context.Context,
	resultsChOut ResultStreamOut[R],
) {
	p.run(ctx, p.private.workersJobsCh, resultsChOut)
}

func (p *WorkerPool[I, R]) run(
	ctx context.Context,
	forwardChOut chan<- Job[I],
	resultsChOut ResultStreamOut[R],
) {
	defer func() {
		close(resultsChOut)
		p.Quit.Done()
		fmt.Printf("<--- WorkerPool.run (QUIT). 🧊🧊🧊\n")
	}()
	fmt.Println("===> 🧊 WorkerPool.run")

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Println("===> 🧊 WorkerPool.run - done received ☢️☢️☢️")
			p.cancelWorkers()

			running = false

		case job, ok := <-p.SourceJobsChIn:
			if ok {
				fmt.Printf("===> 🧊 (#workers: '%v') WorkerPool.run - new job received\n",
					len(p.private.pool),
				)

				if len(p.private.pool) < p.noWorkers {
					p.spawn(ctx, p.private.workersJobsCh, resultsChOut, p.private.finishedCh)
				}
				select {
				case forwardChOut <- job:
					fmt.Printf("===> 🧊 WorkerPool.run - forwarded job 🧿🧿🧿(%v)\n", job.ID)
				case <-ctx.Done(): // ☣️☣️☣️ CHECK THIS, IT MIGHT BE INVALID
					fmt.Printf("===> 🧊 (#workers: '%v') WorkerPool.run - done received ☢️☢️☢️\n",
						len(p.private.pool),
					)
				}
			} else {
				// ⚠️ This close is essential. Since the pool acts as a bridge between
				// 2 channels (p.SourceJobsChIn and p.private.workersJobsCh), when the
				// producer closes p.SourceJobsChIn, we need to delegate that closure
				// to p.private.workersJobsCh, otherwise we end up in a deadlock.
				//
				close(p.private.workersJobsCh)
				fmt.Printf("===> 🚀 WorkerPool.run(source jobs chan closed) 🟥🟥🟥\n")
				running = false
			}
		}
	}

	// We still need to wait for all workers to finish ... Note how we
	// don't pass in the context's Done() channel as it already been consumed
	// in the run loop, and is now closed.
	//
	p.drain(p.private.finishedCh)

	fmt.Printf("===> 🧊 WorkerPool.run - drain complete (workers count: '%v'). 🎃🎃🎃\n",
		len(p.private.pool),
	)
}

func (p *WorkerPool[I, R]) spawn(
	ctx context.Context,
	jobsInCh <-chan Job[I],
	resultsChOut ResultStreamOut[R],
	finishedChOut FinishedStreamOut,
) {
	cancelCh := make(chan CancelWorkSignal, 1)

	w := &workerWrapper[I, R]{
		core: &worker[I, R]{
			id:            p.composeID(),
			fn:            p.fn,
			jobsInCh:      jobsInCh,
			resultsOutCh:  resultsChOut,
			finishedChOut: finishedChOut,
			cancelChIn:    cancelCh,
		},
		cancelChOut: cancelCh,
	}

	p.private.pool[w.core.id] = w
	go w.core.run(ctx)
	fmt.Printf("===> 🧊 WorkerPool.spawned new worker: '%v' 🎀🎀🎀\n", w.core.id)
}

func (p *WorkerPool[I, R]) drain(finishedChIn FinishedStreamIn) {
	fmt.Printf(
		"!!!! 🧊 WorkerPool.drain - waiting for remaining workers: %v (#GRs: %v); 🧊🧊🧊 \n",
		len(p.private.pool), runtime.NumGoroutine(),
	)

	for running := true; running; {
		// 📍 Here, we don't access the finishedChIn channel in a pre-emptive way via
		// the ctx.Done() channel. This is because in a unit test, we define a timeout as
		// part of the test spec using SpecTimeout. When this fires, this is handled by the
		// run loop, which ends that loop then enters drain. When this happens, you can't
		// reuse that same done channel as it will immediately return the value already
		// handled. This has the effect of short-circuiting this loop meaning that
		// workerID := <-finishedChIn never has a chance to be selected and the drain loop
		// exits early. The end result of which means that the p.private.pool collection is
		// never depleted.
		//
		// ⚠️ So an important lesson to be learnt here is that once a ctx.Done() has fired,
		// you can't reuse tha same channel in another select statement as it will simply
		// return immediately, bypassing all the others cases in the select statement.
		//
		// Some noteworthy points:
		//
		// 💎 Safe Access: Accessing the Done() channel concurrently from multiple goroutines
		// is safe. Reading from a closed channel is well-defined behaviour in Go and won't
		// cause panics or issues.
		//
		// 💎 Cancellation Handling: When a context is canceled, the Done() channel is closed,
		// and any goroutine waiting on the channel will be unblocked. Each goroutine needs to
		// have its own select statement to handle the context's cancellation event properly.
		//
		// 💎 Synchronisation: If multiple goroutines are going to react to the context's
		// cancellation, you need to make sure that any shared resources accessed by these
		// goroutines are synchronized properly to avoid race conditions. This might involve
		// using mutexes or other synchronization primitives.
		//
		// 💎 Propagation: If a goroutine creates a child context using context.WithCancel
		// or context.WithTimeout, the child goroutines should use the child context for their
		// operations instead of the parent context. This ensures that the child context's
		// cancellation doesn't affect unrelated goroutines.
		//
		// 💎 Lifetime Management: Be aware of the lifetimes of the contexts and goroutines.
		// If a goroutine outlives its context or keeps references to closed Done() channels,
		// it might not behave as expected.
		//
		workerID := <-finishedChIn
		delete(p.private.pool, workerID)

		if len(p.private.pool) == 0 {
			running = false
		}

		fmt.Printf("!!!! 🧊 WorkerPool.drain - worker(%v) finished, remaining: '%v' 🟥\n",
			workerID, len(p.private.pool),
		)
	}
}

func (p *WorkerPool[I, R]) cancelWorkers() {
	// perhaps, we can replace this with another broadcast mechanism such as sync.Cond
	//
	n := len(p.private.pool)
	for k, w := range p.private.pool {
		fmt.Printf("===> 🧊 cancelling worker '%v' of %v 📛📛📛... \n", k, n)
		// shouldn't need to be preemptable because it is a buffered single item channel
		// which should only ever be accessed by the work pool GR and therefore should
		// never be a position where its competing to send on that channel
		//
		w.cancelChOut <- CancelWorkSignal{}
	}
}
