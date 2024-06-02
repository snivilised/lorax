package boost

import (
	"context"
	"sync"
	"time"

	"github.com/snivilised/lorax/internal/ants"
	"github.com/snivilised/lorax/internal/lo"
)

type (
	ManifoldFunc[I, O any] func(input I) (O, error)
)

type ManifoldFuncPool[I, O any] struct {
	basePool[O]
	functionalPool
	sourceJobsChIn JobStream[I]
}

func NewManifoldFuncPool[I, O any](ctx context.Context,
	size int,
	mf ManifoldFunc[I, O],
	wg *sync.WaitGroup,
	options ...Option,
) (*ManifoldFuncPool[I, O], error) {
	var outputDupCh *Duplex[JobOutput[O]]
	o := ants.LoadOptions(withDefaults(options...)...)
	if o.Output != nil {
		outputDupCh = NewDuplex(make(JobOutputStream[O], o.Output.BufferSize))
	}

	pool, err := ants.NewPoolWithFunc(ctx, size, func(input ants.InputParam) {
		wch := lo.TernaryF(outputDupCh != nil,
			func() JobOutputStreamW[O] {
				return outputDupCh.WriterCh
			},
			func() JobOutputStreamW[O] {
				return nil
			},
		)

		manifoldFuncResponse(ctx, mf, input, wch)
	}, ants.WithOptions(*o))

	return &ManifoldFuncPool[I, O]{
		basePool: basePool[O]{
			wg:          wg,
			outputDupCh: outputDupCh,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}

func (p *ManifoldFuncPool[I, O]) Post(ctx context.Context, input I) error {
	o := p.pool.GetOptions()
	job := Job[I]{
		ID:         o.Generator.Generate(),
		Input:      input,
		SequenceNo: int(p.next()),
	}

	return p.pool.Invoke(ctx, job)
}

func (p *ManifoldFuncPool[I, O]) EndWork(ctx context.Context, interval time.Duration) {
	if p.outputDupCh != nil && !p.ending {
		p.ending = true
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
						close(p.outputDupCh.Channel)
						return
					}
				}
			}
		}(ctx, p, p.wg, interval)
	}
}

func manifoldFuncResponse[I, O any](ctx context.Context,
	mf ManifoldFunc[I, O], input ants.InputParam,
	outputCh JobOutputStreamW[O],
) {
	if job, ok := input.(Job[I]); ok {
		payload, e := mf(job.Input)

		output := JobOutput[O]{
			ID:         job.ID,
			SequenceNo: job.SequenceNo,
			Payload:    payload,
			Error:      e,
		}

		if outputCh != nil {
			select {
			case outputCh <- output:
			// TODO: add a timeout case
			case <-ctx.Done():
			}
		}
	}
}
