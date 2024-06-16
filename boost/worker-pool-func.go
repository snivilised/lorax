package boost

import (
	"context"

	"github.com/snivilised/lorax/internal/ants"
)

type FuncPool[I, O any] struct {
	basePool[I, O]
	functionalPool
	sourceJobsChIn JobStream[I]
}

// NewFuncPool creates a new worker pool using the native ants interface; ie
// new jobs are submitted with Submit(task TaskFunc)
func NewFuncPool[I, O any](ctx context.Context,
	pf ants.PoolFunc,
	wg WaitGroup,
	options ...Option,
) (*FuncPool[I, O], error) {
	// TODO: the automatic invocation of Add/Done might not
	// be valid, need to confirm. I thought that each gr was
	// allocated for each job, but this is not necessarily
	// the case, because each worker has its own job queue.
	//
	pool, err := ants.NewPoolWithFunc(ctx, pf, options...)

	return &FuncPool[I, O]{
		basePool: basePool[I, O]{
			wg: wg,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}
