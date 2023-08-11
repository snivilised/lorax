package async

import (
	"context"
	"fmt"
)

type worker[I any, R any] struct {
	id            WorkerID
	exec          ExecutiveFunc[I, R]
	jobsInCh      JobStreamIn[I]
	resultsOutCh  ResultStreamOut[R]
	finishedChOut FinishedStreamOut

	// this might be better replaced with a broadcast mechanism such as sync.Cond
	//
	cancelChIn CancelStreamIn
}

func (w *worker[I, R]) run(ctx context.Context) {
	defer func() {
		w.finishedChOut <- w.id // ⚠️ non-pre-emptive send, but this should be ok
		fmt.Printf("	<--- 🚀 worker.run(%v) (SENT FINISHED). 🚀🚀🚀\n", w.id)
	}()

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Printf("	---> 🚀 worker.run(%v)(finished) - done received 🔶🔶🔶\n", w.id)

			running = false
		case job, ok := <-w.jobsInCh:
			if ok {
				fmt.Printf("	---> 🚀 worker.run(%v)(input:'%v')\n", w.id, job.Input)
				w.invoke(ctx, job)
			} else {
				fmt.Printf("	---> 🚀 worker.run(%v)(jobs chan closed) 🟥🟥🟥\n", w.id)

				running = false
			}
		}
	}
}

func (w *worker[I, R]) invoke(ctx context.Context, job Job[I]) {
	result, _ := w.exec(job)

	select {
	case <-ctx.Done():
		fmt.Printf("	---> 🚀 worker.invoke(%v)(cancel) - done received 💥💥💥\n", w.id)

	case w.resultsOutCh <- result:
	}
}
