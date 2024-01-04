package boost

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// privateWpInfo (dmz!) contains any state that needs to be mutated in a non concurrent manner
// and therefore should be exclusively accessed by a single go routine. Actually, due to
// our ability to compose functionality with channels as opposed to shared state, the
// pool does not contain any state that is accessed directly or indirectly from other
// go routines. But in the case of the actual core pool, it is mutated without synchronisation
// and hence should only ever be accessed by the worker pool GR in contrast to all the
// other members of WorkerPool. This is an experimental pattern, the purpose of which
// is to clearly indicate what state can be accessed in different concurrency contexts,
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
// is being established for all snivilised projects.
type privateWpInfo[I, O any] struct {
	pool          workersCollection[I, O]
	workersJobsCh chan Job[I]
	finishedCh    finishedStream
	cancelCh      CancelStream
	resultOutCh   PoolResultStreamW
}

// WorkerPool owns the resultOut channel, because it is the only entity that knows
// when all workers have completed their work due to the finished channel, which it also
// owns.
type WorkerPool[I, O any] struct {
	private         privateWpInfo[I, O]
	outputChTimeout time.Duration
	exec            ExecutiveFunc[I, O]
	noWorkers       int
	sourceJobsChIn  JobStream[I]
	RoutineName     GoRoutineName
	WaitAQ          AnnotatedWgAQ
	ResultInCh      PoolResultStreamR
	Logger          *slog.Logger
}

type NewWorkerPoolParams[I, O any] struct {
	NoWorkers       int
	OutputChTimeout time.Duration
	Exec            ExecutiveFunc[I, O]
	JobsCh          JobStream[I]
	CancelCh        CancelStream
	WaitAQ          AnnotatedWgAQ
	Logger          *slog.Logger
}

func NewWorkerPool[I, O any](params *NewWorkerPoolParams[I, O]) *WorkerPool[I, O] {
	noWorkers := runtime.NumCPU()
	if params.NoWorkers > 1 && params.NoWorkers <= MaxWorkers {
		noWorkers = params.NoWorkers
	}

	resultCh := make(PoolResultStream, 1)

	logger := lo.TernaryF(params.Logger == nil,
		func() *slog.Logger {
			return slog.New(zapslog.NewHandler(
				zapcore.NewNopCore(), nil),
			)
		},
		func() *slog.Logger {
			return params.Logger
		},
	)

	wp := &WorkerPool[I, O]{
		private: privateWpInfo[I, O]{
			pool:          make(workersCollection[I, O], noWorkers),
			workersJobsCh: make(JobStream[I], noWorkers),
			finishedCh:    make(finishedStream, noWorkers),
			cancelCh:      params.CancelCh,
			resultOutCh:   resultCh,
		},
		outputChTimeout: params.OutputChTimeout,
		exec:            params.Exec,
		RoutineName:     GoRoutineName("🧊 worker pool"),
		noWorkers:       noWorkers,
		sourceJobsChIn:  params.JobsCh,
		WaitAQ:          params.WaitAQ,
		ResultInCh:      resultCh,
		Logger:          logger,
	}

	return wp
}

// This helps to visualise the activity of the different work threads. Its easier to
// eyeball emojis than worker IDs.
var eyeballs = []string{
	"❤️", "💙", "💚", "💜", "💛", "🤍", "💖", "💗", "💝",
}

func (p *WorkerPool[I, O]) composeID() workerID {
	n := len(p.private.pool)
	index := (n) % len(eyeballs)
	emoji := eyeballs[index]

	return workerID(fmt.Sprintf("(%v)WORKER-ID-%v:%v", emoji, n, uuid.NewString()))
}

func (p *WorkerPool[I, O]) Start(
	parentContext context.Context,
	parentCancel context.CancelFunc,
	outputsChOut chan<- JobOutput[O],
) {
	p.run(parentContext,
		parentCancel,
		p.outputChTimeout,
		p.private.workersJobsCh,
		outputsChOut,
	)
}

func (p *WorkerPool[I, O]) run(
	parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
	forwardChOut JobStreamW[I],
	outputsChOut JobOutputStreamW[O],
) {
	result := &PoolResult{}
	defer func(r *PoolResult) {
		if outputsChOut != nil {
			close(outputsChOut)
		}
		p.private.resultOutCh <- r

		p.WaitAQ.Done(p.RoutineName)
		p.Logger.Debug("<--- WorkerPool.run (QUIT). 🧊🧊🧊\n")
	}(result)
	p.Logger.Debug(fmt.Sprintf(
		"===> 🧊 WorkerPool.run ...(ctx:%+v)\n",
		parentContext,
	))

	for running := true; running; {
		select {
		case <-parentContext.Done():
			running = false

			close(forwardChOut) // ⚠️ This is new
			p.Logger.Debug("===> 🧊 WorkerPool.run (source jobs chan closed) - done received ☢️☢️☢️")

		case job, ok := <-p.sourceJobsChIn:
			if ok {
				p.Logger.Debug(fmt.Sprintf(
					"===> 🧊 (#workers: '%v') WorkerPool.run - new job received",
					len(p.private.pool),
				))

				if len(p.private.pool) < p.noWorkers {
					p.spawn(parentContext,
						parentCancel,
						outputChTimeout,
						p.private.workersJobsCh,
						outputsChOut,
						p.private.finishedCh,
					)
				}
				select {
				case forwardChOut <- job:
					p.Logger.Debug(fmt.Sprintf(
						"===> 🧊 WorkerPool.run - forwarded job 🧿🧿🧿(%v) [Seq: %v]",
						job.ID,
						job.SequenceNo,
					))
				case <-parentContext.Done():
					running = false

					close(forwardChOut) // ⚠️ This is new
					p.Logger.Debug(fmt.Sprintf(
						"===> 🧊 (#workers: '%v') WorkerPool.run - done received ☢️☢️☢️",
						len(p.private.pool),
					))
				}
			} else {
				// ⚠️ This close is essential. Since the pool acts as a bridge between
				// 2 channels (p.sourceJobsChIn and p.private.workersJobsCh/forwardChOut),
				// when the producer closes p.sourceJobsChIn, we need to delegate that
				// closure to forwardChOut, otherwise we end up in a deadlock.
				//
				running = false
				close(forwardChOut)
				p.Logger.Debug("===> 🚀 WorkerPool.run(source jobs chan closed) 🟥🟥🟥")
			}
		}
	}

	// We still need to wait for all workers to finish ... Note how we
	// don't pass in the context's Done() channel as it already been consumed
	// in the run loop, and is now closed.
	//
	if err := p.drain(p.private.finishedCh); err != nil {
		result.Error = err

		p.Logger.Debug(fmt.Sprintf(
			"===> 🧊 WorkerPool.run - drain complete with error: '%v' (workers count: '%v'). 📛📛📛",
			err,
			len(p.private.pool),
		))
	} else {
		p.Logger.Debug(fmt.Sprintf(
			"===> 🧊 WorkerPool.run - drain complete OK (workers count: '%v'). ☑️☑️☑️",
			len(p.private.pool),
		))
	}
}

func (p *WorkerPool[I, O]) spawn(
	parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
	jobsChIn JobStreamR[I],
	outputsChOut JobOutputStreamW[O],
	finishedChOut finishedStreamW,
) {
	w := &workerWrapper[I, O]{
		core: &worker[I, O]{
			id:            p.composeID(),
			exec:          p.exec,
			jobsChIn:      jobsChIn,
			outputsChOut:  outputsChOut,
			finishedChOut: finishedChOut,
			logger:        p.Logger,
		},
	}

	p.private.pool[w.core.id] = w
	go w.core.run(parentContext, parentCancel, outputChTimeout)
	p.Logger.Debug(fmt.Sprintf(
		"===> 🧊 WorkerPool.spawned new worker: '%v' 🎀🎀🎀",
		w.core.id,
	))
}

func (p *WorkerPool[I, O]) drain(finishedChIn finishedStreamR) error {
	p.Logger.Debug(fmt.Sprintf(
		"!!!! 🧊 WorkerPool.drain - waiting for remaining workers: %v (#GRs: %v); 🧊🧊🧊",
		len(p.private.pool), runtime.NumGoroutine(),
	))

	var firstError error

	for running := true; running; {
		// 📍 Here, we don't access the finishedChIn channel in a pre-emptive way via
		// the parentContext.Done() channel. This is because in a unit test, we define a timeout as
		// part of the test spec using SpecTimeout. When this fires, this is handled by the
		// run loop, which ends that loop then enters drain the phase. When this happens,
		// you can't reuse that same done channel as it will immediately return the value
		// already handled. This has the effect of short-circuiting this loop meaning that
		// workerResult := <-finishedChIn never has a chance to be selected and the drain loop
		// exits early. The end result of which means that the p.private.pool collection is
		// never depleted.
		//
		// ⚠️ So an important lesson to be learnt here is that once a parentContext.Done() has fired,
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
		workerResult := <-finishedChIn
		delete(p.private.pool, workerResult.id)

		if len(p.private.pool) == 0 {
			running = false
		}

		if workerResult.err != nil {
			p.Logger.Debug(fmt.Sprintf(
				"!!!! 🧊 WorkerPool.drain - worker (%v) 💢💢💢 finished with error: '%v'",
				workerResult.id,
				workerResult.err,
			))

			if firstError == nil {
				firstError = workerResult.err
			}
		}

		p.Logger.Debug(fmt.Sprintf(
			"!!!! 🧊 WorkerPool.drain - worker-result-error(%v) finished, remaining: '%v' 🟥",
			workerResult.err, len(p.private.pool),
		))
	}

	return firstError
}
