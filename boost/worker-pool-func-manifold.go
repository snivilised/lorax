package boost

import (
	"context"
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type ManifoldFunc ants.PoolFunc

type ManifoldFuncPool[I, O any] struct {
	basePool
	functionalPool
	sourceJobsChIn JobStream[I]
}

func NewManifoldFuncPool[I, O any](ctx context.Context,
	size int,
	pf ants.PoolFunc,
	wg *sync.WaitGroup,
	options ...Option,
) (*ManifoldFuncPool[I, O], error) {
	pool, err := ants.NewPoolWithFunc(ctx, size, func(i ants.InputParam) {
		defer wg.Done()

		pf(i)
	}, options...)

	return &ManifoldFuncPool[I, O]{
		basePool: basePool{
			ctx:   ctx,
			idGen: &Sequential{},
			wg:    wg,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}
