package boost

import (
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type WorkerPoolInvoker[I, O any] struct {
	basePool
	functionalPool
	sourceJobsChIn JobStream[I]
}

// NewFuncPool creates a new worker pool using the native ants interface; ie
// new jobs are submitted with Submit(task TaskFunc)
func NewFuncPool[I, O any](size int,
	pf ants.PoolFunc,
	wg *sync.WaitGroup,
	options ...Option,
) (*WorkerPoolInvoker[I, O], error) {
	// TODO: the automatic invocation of Add/Done might not
	// be valid, need to confirm. I thought that each gr was
	// allocated for each job, but this is not necessarily
	// the case, because each worker has its own job queue.
	//
	pool, err := ants.NewPoolWithFunc(size, func(i ants.InputParam) {
		defer wg.Done()
		pf(i)
	}, options...)

	return &WorkerPoolInvoker[I, O]{
		basePool: basePool{
			idGen: &Sequential{},
			wg:    wg,
		},
		functionalPool: functionalPool{
			pool: pool,
		},
	}, err
}

func (p *WorkerPoolInvoker[I, O]) Post(job ants.InputParam) error {
	p.wg.Add(1) // because the gr lifetime is tied to the job not the worker

	return p.pool.Invoke(job)
}

func (p *WorkerPoolInvoker[I, O]) Running() int {
	return p.pool.Running()
}

func (p *WorkerPoolInvoker[I, O]) Release() {
	p.pool.Release()
}
