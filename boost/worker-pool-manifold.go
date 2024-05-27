package boost

import (
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type manifoldFunc ants.PoolFunc

type WorkerPoolManifold[I, O any] struct {
	basePool
	functionalPool
	sourceJobsChIn JobStream[I]
}

func NewManifoldPool[I, O any](size int,
	pf ants.PoolFunc,
	wg *sync.WaitGroup,
	options ...Option,
) (*WorkerPoolManifold[I, O], error) {
	pool, err := ants.NewPoolWithFunc(size, func(i ants.InputParam) {
		defer wg.Done()

		pf(i)
	}, options...)

	return &WorkerPoolManifold[I, O]{
		basePool: basePool{
			idGen: &Sequential{},
			wg:    wg,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}
