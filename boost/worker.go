package boost

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type worker[I any, O any] struct {
	id            workerID
	exec          ExecutiveFunc[I, O]
	jobsChIn      JobStreamR[I]
	outputsChOut  OutputStream[O]
	finishedChOut finishedStreamW
}

func (w *worker[I, O]) run(parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
) {
	result := workerFinishedResult{
		id: w.id,
	}
	defer func(r *workerFinishedResult) {
		w.finishedChOut <- r // ⚠️ non-pre-emptive send, but this should be ok

		fmt.Printf("	<--- 🚀 worker.run(%v) (SENT FINISHED - error:'%v'). 🚀🚀🚀\n", w.id, r.err)
	}(&result)

	fmt.Printf("	---> 🚀 worker.run(%v) ...(ctx:%+v)\n", w.id, parentContext)

	for running := true; running; {
		select {
		case <-parentContext.Done():
			fmt.Printf("	---> 🚀 worker.run(%v)(finished) - done received 🔶🔶🔶\n", w.id)

			running = false
		case job, ok := <-w.jobsChIn:
			if ok {
				fmt.Printf("	---> 🚀 worker.run(%v)(input:'%v')\n", w.id, job.Input)
				err := w.invoke(parentContext, parentCancel, outputChTimeout, job)

				if err != nil {
					result.err = err
					running = false
				}
			} else {
				fmt.Printf("	---> 🚀 worker.run(%v)(jobs chan closed) 🟥🟥🟥\n", w.id)

				running = false
			}
		}
	}
}

func (w *worker[I, O]) invoke(parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
	job Job[I],
) error {
	var err error

	outputContext, cancel := context.WithTimeout(parentContext, outputChTimeout)
	defer cancel()

	result, _ := w.exec(job)

	if w.outputsChOut != nil {
		fmt.Printf("	---> 🚀 worker.invoke ⏰ output timeout: '%v'\n", outputChTimeout)

		select {
		case w.outputsChOut <- result:

		case <-parentContext.Done():
			fmt.Printf("	---> 🚀 worker.invoke(%v)(cancel) - done received 💥💥💥\n", w.id)

		case <-outputContext.Done():
			fmt.Printf("	---> 🚀 worker.invoke(%v)(cancel) - timeout on send 👿👿👿\n", w.id)

			// ??? err = i18n.NewOutputChTimeoutError()
			err = errors.New("timeout on send")

			parentCancel()
		}
	}

	return err
}
