package boost

// the Job is implied by the I and O specified, so
// Job[I] is derived from WorkerPool[I]
// JobOutput[O] is derived from WorkerPool[O]
//
// Unit of Work schemes:
//
// * Simple: this is the native ants format
// Simple: func()
// SimpleE: func() error
//
// * Input: input only
// Input: func[I any]()
// InputE: func[I any]() error
//
// * Output: output only, is this really useful?
// Output: func[O any]() O
// OutputE: func[O any]() O, error
//
// * Manifold: with input and output
// Manifold: func[I, O any]() O
// ManifoldE: func[I, O any]() O, error
//
import (
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type WorkerPool[I, O any] struct {
	basePool
	generalPool
	sourceJobsChIn JobStream[I]
}

// NewSubmitterPool creates a new worker pool using the native ants interface; ie
// new jobs are submitted with Submit(task TaskFunc)
func NewSubmitterPool[I, O any](size int,
	wg *sync.WaitGroup,
	options ...Option,
) (*WorkerPool[I, O], error) {
	pool, err := ants.NewPool(size, options...)

	return &WorkerPool[I, O]{
		basePool: basePool{
			idGen: &Sequential{},
			wg:    wg,
		},
		generalPool: generalPool{
			pool: pool,
		},
	}, err
}

func (p *WorkerPool[I, O]) Post(task ants.TaskFunc) error {
	p.wg.Add(1)

	return p.pool.Submit(func() {
		defer p.wg.Done()

		task()
	})
}

func (p *WorkerPool[I, O]) Release() {
	p.pool.Release()
}
