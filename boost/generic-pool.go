package boost

import (
	"context"
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

// functionalPool
type functionalPool struct {
	pool *ants.PoolWithFunc
}

func (p *functionalPool) Post(ctx context.Context, job InputParam) error {
	return p.pool.Invoke(ctx, job)
}

func (p *functionalPool) Release(ctx context.Context) {
	p.pool.Release(ctx)
}

func (p *functionalPool) Running() int {
	return p.pool.Running()
}

func (p *functionalPool) Waiting() int {
	return p.pool.Waiting()
}

// taskPool
type taskPool struct {
	pool *ants.Pool
}

func (p *taskPool) Post(ctx context.Context, task TaskFunc) error {
	return p.pool.Submit(ctx, task)
}

func (p *taskPool) Release(ctx context.Context) {
	p.pool.Release(ctx)
}

func (p *taskPool) Running() int {
	return p.pool.Running()
}

func (p *taskPool) Waiting() int {
	return p.pool.Waiting()
}

func source[I any](ctx context.Context,
	wg *sync.WaitGroup, o *ants.Options,
	injectable injectable[I],
	closable closable,
) *Duplex[I] {
	inputDupCh := NewDuplex(make(SourceStream[I], o.Input.BufferSize))

	wg.Add(1)
	go func(ctx context.Context) { //nolint:wsl // pedant
		defer func() {
			closable.terminate()
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case input, ok := <-inputDupCh.ReaderCh:
				if !ok {
					return
				}

				_ = injectable.inject(input)
			}
		}
	}(ctx)

	return inputDupCh
}
