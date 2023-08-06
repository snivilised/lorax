package async

import (
	"context"
	"fmt"
)

type workersCollection[I, R any] map[WorkerID]*worker[I, R]

type worker[I any, R any] struct {
	id WorkerID

	// TODO: there is still no benefit on using an interface rather than a function,
	// might have to change this back to a function
	//
	fn Executive[I, R]
}

func (w *worker[I, R]) accept(ctx context.Context, info *workerInfo[I, R]) {
	fmt.Printf("---> 🚀 worker.accept: '%v', input:'%v'\n", w.id, info.job.Input)
	result, _ := w.fn.Invoke(info.job)

	select { // BREAKS: when cancellation occurs, send on closed chan
	case <-ctx.Done():
		fmt.Println("---> 🚀 worker.accept(result) - done received 💥💥💥")

	case info.resultsOut <- result:
	}

	select {
	case <-ctx.Done():
		fmt.Println("---> 🚀 worker.accept(finished) - done received ❌❌❌")

	case info.finishedOut <- w.id:
	}
}
