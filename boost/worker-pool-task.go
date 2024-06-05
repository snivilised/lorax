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
	"context"
	"sync"

	"github.com/snivilised/lorax/internal/ants"
)

type TaskPool[I, O any] struct {
	basePool[I, O]
	taskPool
	sourceJobsChIn JobStream[I]
}

// NewTaskPool creates a new worker pool using the native ants interface; ie
// new jobs are submitted with Submit(task TaskFunc)
func NewTaskPool[I, O any](ctx context.Context,
	size int,
	wg *sync.WaitGroup,
	options ...Option,
) (*TaskPool[I, O], error) {
	pool, err := ants.NewPool(ctx, size, withDefaults(options...)...)

	return &TaskPool[I, O]{
		basePool: basePool[I, O]{
			wg: wg,
		},
		taskPool: taskPool{
			pool: pool,
		},
	}, err
}
