package boost

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/snivilised/lorax/internal/lo"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// privateWpInfoL (dmz!) contains any state that needs to be mutated in a non concurrent manner
// and therefore should be exclusively accessed by a single go routine. Actually, due to
// our ability to compose functionality with channels as opposed to shared state, the
// pool does not contain any state that is accessed directly or indirectly from other
// go routines. But in the case of the actual core pool, it is mutated without synchronisation
// and hence should only ever be accessed by the worker pool GR in contrast to all the
// other members of WorkerPool. This is an experimental pattern, the purpose of which
// is to clearly indicate what state can be accessed in different concurrency contexts,
// to ensure future updates can be applied with minimal cognitive overload.
//
// There is another purpose for privateWpInfoL and that is to do with "confinement" as
// described on page 86 of CiG. The aim here is to use "lexical confinement" for
// duplex channel definitions, so although a channel is thread safe so ordinarily
// would not be a candidate member of privateWpInfoL, a duplex channel ought to be
// protected from accidentally being used incorrectly, ie trying to write to a channel
// that is meant to be read only. So methods that use a channel should now receive the
// channel through a method parameter (defined as either chan<-, or <-chan), rather
// than be expected to simply access the member variable directly. This clearly signals
// that any channel defined in privateWpInfoL should never to accessed directly (other
// than for passing it to another method). This is an experimental convention that
// is being established for all snivilised projects.
type privateWpInfoL[I, O any] struct {
	pool          workersCollectionL[I, O]
	workersJobsCh chan Job[I]
	finishedCh    finishedStream
	cancelCh      CancelStream
	resultOutCh   PoolResultStreamW
}

// WorkerPoolL owns the resultOut channel, because it is the only entity that knows
// when all workers have completed their work due to the finished channel, which it also
// owns.
// Deprecated: use ants base worker-pool instead.
type WorkerPoolL[I, O any] struct {
	private         privateWpInfoL[I, O]
	outputChTimeout time.Duration
	exec            ExecutiveFunc[I, O]
	noWorkers       int
	sourceJobsChIn  JobStream[I]
	RoutineName     GoRoutineName
	WaitAQ          AnnotatedWgAQ
	ResultInCh      PoolResultStreamR
	Logger          *slog.Logger
}

// NewWorkerPoolParamsL
// Deprecated: use ants base worker-pool instead.
type NewWorkerPoolParamsL[I, O any] struct {
	NoWorkers       int
	OutputChTimeout time.Duration
	Exec            ExecutiveFunc[I, O]
	JobsCh          JobStream[I]
	CancelCh        CancelStream
	WaitAQ          AnnotatedWgAQ
	Logger          *slog.Logger
}

// NewWorkerPoolL
// Deprecated: use ants base worker-pool instead.
func NewWorkerPoolL[I, O any](params *NewWorkerPoolParamsL[I, O]) *WorkerPoolL[I, O] {
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

	wp := &WorkerPoolL[I, O]{
		private: privateWpInfoL[I, O]{
			pool:          make(workersCollectionL[I, O], noWorkers),
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

func (p *WorkerPoolL[I, O]) composeID() workerID {
	n := len(p.private.pool)
	index := (n) % len(eyeballs)
	emoji := eyeballs[index]

	return workerID(fmt.Sprintf("(%v)WORKER-ID-%v:%v", emoji, n, uuid.NewString()))
}

func (p *WorkerPoolL[I, O]) Start(
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

func (p *WorkerPoolL[I, O]) run(
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
	}(result)

	for running := true; running; {
		select {
		case <-parentContext.Done():
			running = false

			close(forwardChOut)
			p.Logger.Debug("source jobs chan closed - done received",
				slog.String("source", "worker-pool.run"),
			)

		case job, ok := <-p.sourceJobsChIn:
			if ok {
				p.Logger.Debug("new job received",
					slog.String("source", "worker-pool.run"),
					slog.Int("pool size", len(p.private.pool)),
				)

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
					p.Logger.Debug("forwarded job",
						slog.String("source", "worker-pool.run"),
						slog.String("job-id", job.ID),
						slog.Int("sequence-no", job.SequenceNo),
					)
				case <-parentContext.Done():
					running = false

					close(forwardChOut)
					p.Logger.Debug("done received",
						slog.String("source", "worker-pool.run"),
						slog.Int("pool size", len(p.private.pool)),
					)
				}
			} else {
				// 💫 This close is essential. Since the pool acts as a bridge between
				// 2 channels (p.sourceJobsChIn and p.private.workersJobsCh/forwardChOut),
				// when the producer closes p.sourceJobsChIn, we need to delegate that
				// closure to forwardChOut, otherwise we end up in a deadlock.
				//
				running = false

				close(forwardChOut)

				p.Logger.Debug("source jobs chan closed",
					slog.String("source", "worker-pool.run"),
				)
			}
		}
	}

	// We still need to wait for all workers to finish ... Note how we
	// don't pass in the context's Done() channel as it already been consumed
	// in the run loop, and is now closed.
	//
	if err := p.drain(p.private.finishedCh); err != nil {
		result.Error = err

		p.Logger.Error("drain complete with error",
			slog.String("source", "worker-pool.run"),
			slog.Int("pool size", len(p.private.pool)),
			slog.String("error", err.Error()),
		)
	} else {
		p.Logger.Debug("drain complete OK",
			slog.String("source", "worker-pool.run"),
			slog.Int("pool size", len(p.private.pool)),
		)
	}
}

func (p *WorkerPoolL[I, O]) spawn(
	parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
	jobsChIn JobStreamR[I],
	outputsChOut JobOutputStreamW[O],
	finishedChOut finishedStreamW,
) {
	w := &workerWrapperL[I, O]{
		core: &workerL[I, O]{
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
	p.Logger.Debug("spawned new worker",
		slog.String("source", "WorkerPool.spawn"),
		slog.String("worker-id", string(w.core.id)),
	)
}

func (p *WorkerPoolL[I, O]) drain(finishedChIn finishedStreamR) error {
	p.Logger.Debug("waiting for remaining workers...",
		slog.String("source", "WorkerPool.drain"),
		slog.Int("pool size", len(p.private.pool)),
	)

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
			p.Logger.Error("worker finished with error",
				slog.String("source", "WorkerPool.drain"),
				slog.String("result-id", string(workerResult.id)),
				slog.String("error", workerResult.err.Error()),
			)

			if firstError == nil {
				firstError = workerResult.err
			}
		}

		p.Logger.Debug("worker pool finished",
			slog.String("source", "WorkerPool.drain"),
			slog.Int("remaining", len(p.private.pool)),
		)
	}

	return firstError
}
