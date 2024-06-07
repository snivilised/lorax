package boost

import (
	"context"
	"sync"
	"time"

	"github.com/snivilised/lorax/internal/ants"
)

type (
	// ManifoldFunc is the pre-defined function registered with the worker
	// pool, executed for each incoming job.
	ManifoldFunc[I, O any] func(input I) (O, error)
)

// ManifoldFuncPool is a wrapper around the underlying ants function based
// worker pool. The client is expected to create an output channel to
// receive the outputs of executing jobs in the worker pool. If the
// output channel is not defined, then jobs will still be executed, but
// the output of which will not be sent, also losing job execution error
// status.
type ManifoldFuncPool[I, O any] struct {
	basePool[I, O]
	functionalPool
}

// NewManifoldFuncPool creates a new manifold function based worker pool.
func NewManifoldFuncPool[I, O any](ctx context.Context,
	mf ManifoldFunc[I, O],
	wg *sync.WaitGroup,
	options ...Option,
) (*ManifoldFuncPool[I, O], error) {
	var (
		oi *outputInfo[O]
		wi *outputInfoW[O]
		o  = ants.LoadOptions(withDefaults(options...)...)
	)

	if oi = newOutputInfo[O](o); oi != nil {
		wi = fromOutputInfo(o, oi)
	}

	pool, err := ants.NewPoolWithFunc(ctx, func(input InputParam) {
		manifoldFuncResponse(ctx, mf, input, wi)
	}, ants.WithOptions(*o))

	return &ManifoldFuncPool[I, O]{
		basePool: basePool[I, O]{
			wg: wg,
			oi: oi,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}

// Post allows the client to submit to the work pool represented by
// input values of type I.
func (p *ManifoldFuncPool[I, O]) Post(ctx context.Context, input I) error {
	o := p.pool.GetOptions()
	job := Job[I]{
		ID:         o.Generator.Generate(),
		Input:      input,
		SequenceNo: int(p.next()),
	}

	return p.pool.Invoke(ctx, job)
}

// Source returns an input stream through which the client can submit
// jobs to the pool. Using an input stream vs invoking Post is
// mutually exclusive; that is to say, if Source is called, then Post
// must not be called; any such invocations will be ignored.
func (p *ManifoldFuncPool[I, O]) Source(ctx context.Context,
	wg *sync.WaitGroup,
) SourceStreamW[I] {
	o := p.pool.GetOptions()

	p.basePool.inputDupCh = source(ctx, wg, o,
		injector[I](func(input I) error {
			return p.Post(ctx, input)
		}),
		terminator(func() {
			p.Conclude(ctx)
		}),
	)

	return p.basePool.inputDupCh.WriterCh
}

// Conclude signifies to the worker pool that no more work will be
// submitted. When submitting to the pool directly using the
// Post method, the client must call this method. Failure to do so
// will result in a pool that never ends. When the client elects
// to use an input channel, by invoking Source, then Conclude will
// be called automatically as long as the input channel has been closed.
// Failure to close the channel will again result in a never ending
// worker pool.
func (p *ManifoldFuncPool[I, O]) Conclude(ctx context.Context) {
	if p.oi != nil && !p.ending {
		p.ending = true
		o := p.pool.GetOptions()
		interval := max(o.Output.CheckCloseInterval, ants.MinimumCheckCloseInterval)

		p.wg.Add(1)
		go func(ctx context.Context,
			pool *ManifoldFuncPool[I, O],
			wg *sync.WaitGroup,
			interval time.Duration,
		) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case <-time.After(interval):
					if pool.Running() == 0 && pool.Waiting() == 0 {
						close(p.oi.outputDupCh.Channel)
						return
					}
				}
			}
		}(ctx, p, p.wg, interval)
	}
}

func manifoldFuncResponse[I, O any](ctx context.Context,
	mf ManifoldFunc[I, O],
	input InputParam,
	wi *outputInfoW[O],
) {
	if job, ok := input.(Job[I]); ok {
		payload, e := mf(job.Input)

		output := JobOutput[O]{
			ID:         job.ID,
			SequenceNo: job.SequenceNo,
			Payload:    payload,
			Error:      e,
		}

		if wi != nil {
			_ = respond(ctx, wi, &output)
		}
	}
}
