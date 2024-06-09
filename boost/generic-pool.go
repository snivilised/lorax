package boost

import (
	"context"
	"errors"
	"time"

	"github.com/snivilised/lorax/internal/ants"
	"github.com/snivilised/lorax/internal/lo"
)

// functionalPool
type functionalPool struct {
	pool *ants.PoolWithFunc
}

// Post submits a task to the pool.
func (p *functionalPool) Post(ctx context.Context, job InputParam) error {
	return p.pool.Invoke(ctx, job)
}

// Release closes this pool and releases the worker queue.
func (p *functionalPool) Release(ctx context.Context) {
	p.pool.Release(ctx)
}

// Running returns the number of workers currently running.
func (p *functionalPool) Running() int {
	return p.pool.Running()
}

// Waiting returns the number of tasks waiting to be executed.
func (p *functionalPool) Waiting() int {
	return p.pool.Waiting()
}

// taskPool
type taskPool struct {
	pool *ants.Pool
}

// Post submits a task to the pool.
func (p *taskPool) Post(ctx context.Context, task TaskFunc) error {
	return p.pool.Submit(ctx, task)
}

// Release closes this pool and releases the worker queue.
func (p *taskPool) Release(ctx context.Context) {
	p.pool.Release(ctx)
}

// Running returns the number of workers currently running.
func (p *taskPool) Running() int {
	return p.pool.Running()
}

// Waiting returns the number of tasks waiting to be executed.
func (p *taskPool) Waiting() int {
	return p.pool.Waiting()
}

func source[I any](ctx context.Context,
	wg WaitGroup, o *ants.Options,
	injectable injectable[I],
	closable closable,
) *Duplex[I] {
	inputDupCh := NewDuplex(make(SourceStream[I], o.Input.BufferSize))

	wg.Add(1)
	go func(ctx context.Context, inputCh SourceStreamR[I]) {
		defer func() {
			closable.terminate()
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case input, ok := <-inputCh:
				if !ok {
					return
				}

				_ = injectable.inject(input)
			}
		}
	}(ctx, inputDupCh.ReaderCh)

	return inputDupCh
}

func newOutputInfo[O any](o *Options) *outputInfo[O] {
	if o.Output == nil {
		return nil
	}

	return &outputInfo[O]{
		cancelDupCh: NewDuplex(make(CancelStream, 1)),
		outputDupCh: NewDuplex(make(JobOutputStream[O], o.Output.BufferSize)),
	}
}

// fromOutputs assumes o.Output is defined
func fromOutputInfo[O any](o *Options, oi *outputInfo[O]) *outputInfoW[O] {
	const never = time.Hour * 50000

	timeout := lo.TernaryF(o.Output != nil,
		func() time.Duration {
			return max(o.Output.TimeoutOnSend, ants.MinimumTimeoutOnSend)
		},
		func() time.Duration {
			return never
		},
	)

	return &outputInfoW[O]{
		cancelCh:      oi.cancelDupCh.WriterCh,
		outputCh:      oi.outputDupCh.WriterCh,
		timeoutOnSend: timeout,
	}
}

func respond[O any](ctx context.Context, wi *outputInfoW[O], output *JobOutput[O]) (err error) {
	select {
	case wi.outputCh <- *output:
		return nil
	case <-time.After(wi.timeoutOnSend):
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case wi.cancelCh <- CancelWorkSignal{}:
			err = errors.New("timeout")
		}

	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}
