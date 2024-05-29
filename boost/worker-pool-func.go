package boost

import (
	"context"
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type FuncPool[I, O any] struct {
	basePool
	functionalPool
	sourceJobsChIn JobStream[I]
}

// NewFuncPool creates a new worker pool using the native ants interface; ie
// new jobs are submitted with Submit(task TaskFunc)
func NewFuncPool[I, O any](ctx context.Context,
	size int,
	pf ants.PoolFunc,
	wg *sync.WaitGroup,
	options ...Option,
) (*FuncPool[I, O], error) {
	// TODO: the automatic invocation of Add/Done might not
	// be valid, need to confirm. I thought that each gr was
	// allocated for each job, but this is not necessarily
	// the case, because each worker has its own job queue.
	//
	pool, err := ants.NewPoolWithFunc(ctx, size, pf, options...)

	return &FuncPool[I, O]{
		basePool: basePool{
			ctx:   ctx,
			wg:    wg,
			idGen: &Sequential{},
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}

func (p *FuncPool[I, O]) Post(job ants.InputParam) error {
	return p.pool.Invoke(job)
}

func (p *FuncPool[I, O]) Running() int {
	return p.pool.Running()
}

func (p *FuncPool[I, O]) Release() {
	p.pool.Release()
}
