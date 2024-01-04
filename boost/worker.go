package boost

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type worker[I any, O any] struct {
	id            workerID
	exec          ExecutiveFunc[I, O]
	jobsChIn      JobStreamR[I]
	outputsChOut  JobOutputStreamW[O]
	finishedChOut finishedStreamW
	logger        *slog.Logger
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

		w.logger.Debug(fmt.Sprintf("	<--- 🚀 worker.run(%v) (SENT FINISHED - error:'%v'). 🚀🚀🚀",
			w.id, r.err,
		))
	}(&result)

	w.logger.Debug(fmt.Sprintf("	---> 🚀 worker.run(%v) ...(ctx:%+v)\n", w.id, parentContext))

	for running := true; running; {
		select {
		case <-parentContext.Done():
			w.logger.Debug(fmt.Sprintf(
				"	---> 🚀 worker.run(%v)(finished) - done received 🔶🔶🔶", w.id,
			))

			running = false
		case job, ok := <-w.jobsChIn:
			if ok {
				w.logger.Debug(fmt.Sprintf(
					"	---> 🚀 worker.run(%v)(input:'%v')", w.id, job.Input,
				))

				err := w.invoke(parentContext, parentCancel, outputChTimeout, job)

				if err != nil {
					result.err = err
					running = false
				}
			} else {
				w.logger.Debug(fmt.Sprintf(
					"	---> 🚀 worker.run(%v)(jobs chan closed) 🟥🟥🟥", w.id,
				))

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
		w.logger.Debug(fmt.Sprintf(
			"	---> 🚀 worker.invoke ⏰ output timeout: '%v'", outputChTimeout,
		))

		select {
		case w.outputsChOut <- result:

		case <-parentContext.Done():
			w.logger.Debug(fmt.Sprintf(
				"	---> 🚀 worker.invoke(%v)(cancel) - done received 💥💥💥", w.id,
			))

		case <-outputContext.Done():
			w.logger.Debug(fmt.Sprintf(
				"	---> 🚀 worker.invoke(%v)(cancel) - timeout on send 👿👿👿", w.id,
			))

			// ??? err = i18n.NewOutputChTimeoutError()
			err = errors.New("timeout on send")

			parentCancel()
		}
	}

	return err
}
