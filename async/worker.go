package async

import (
	"context"
	"fmt"
)

type worker[I any, O any] struct {
	id            WorkerID
	exec          ExecutiveFunc[I, O]
	jobsChIn      JobStreamR[I]
	outputsChOut  OutputStream[O]
	finishedChOut FinishedStreamW

	// this might be better replaced with a broadcast mechanism such as sync.Cond
	//
	cancelChIn CancelStreamR
}

func (w *worker[I, O]) run(ctx context.Context) {
	defer func() {
		w.finishedChOut <- w.id // ⚠️ non-pre-emptive send, but this should be ok
		fmt.Printf("	<--- 🚀 worker.run(%v) (SENT FINISHED). 🚀🚀🚀\n", w.id)
	}()
	fmt.Printf("	---> 🚀 worker.run(%v) ...(ctx:%+v)\n", w.id, ctx)

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Printf("	---> 🚀 worker.run(%v)(finished) - done received 🔶🔶🔶\n", w.id)

			running = false
		case job, ok := <-w.jobsChIn:
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

func (w *worker[I, O]) invoke(ctx context.Context, job Job[I]) {
	result, _ := w.exec(job)

	if w.outputsChOut != nil {
		select {
		case <-ctx.Done():
			fmt.Printf("	---> 🚀 worker.invoke(%v)(cancel) - done received 💥💥💥\n", w.id)

		case w.outputsChOut <- result:
		}
	}
}
