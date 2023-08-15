package async

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/google/uuid"
)

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
	exec           ExecutiveFunc[I, R]
	noWorkers      int
	SourceJobsChIn JobStreamIn[I]

	Quit *sync.WaitGroup
}

type NewWorkerPoolParams[I, R any] struct {
	NoWorkers int
	Exec      ExecutiveFunc[I, R]
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
		exec:           params.Exec,
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
	fmt.Printf("===> 🧊 WorkerPool.run ...(ctx:%+v)\n", ctx)

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Println("===> 🧊 WorkerPool.run - done received ☢️☢️☢️")

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
					fmt.Printf("===> 🧊 WorkerPool.run - forwarded job 🧿🧿🧿(%v) [Seq: %v]\n",
						job.ID,
						job.SequenceNo,
					)
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
	jobsChIn JobStreamIn[I],
	resultsChOut ResultStreamOut[R],
	finishedChOut FinishedStreamOut,
) {
	cancelCh := make(chan CancelWorkSignal, 1)

	w := &workerWrapper[I, R]{
		core: &worker[I, R]{
			id:            p.composeID(),
			exec:          p.exec,
			jobsChIn:      jobsChIn,
			resultsChOut:  resultsChOut,
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
		// run loop, which ends that loop then enters drain the phase. When this happens,
		// you can't reuse that same done channel as it will immediately return the value
		// already handled. This has the effect of short-circuiting this loop meaning that
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
