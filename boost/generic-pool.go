package boost

import (
	"context"

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
