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
		w.finishedChOut <- r // 💫 non-pre-emptive send, but this should be ok

		w.logger.Debug("send complete",
			slog.String("source", "worker.run"),
			slog.String("worker-id", string(w.id)),
		)
	}(&result)

	for running := true; running; {
		select {
		case <-parentContext.Done():
			w.logger.Debug("finished - done received",
				slog.String("source", "worker.run"),
				slog.String("worker-id", string(w.id)),
			)

			running = false
		case job, ok := <-w.jobsChIn:
			if ok {
				w.logger.Debug("read from channel",
					slog.String("source", "worker.run"),
					slog.String("worker-id", string(w.id)),
					slog.String("job-input", fmt.Sprintf("%v", w.id)),
				)

				err := w.invoke(parentContext, parentCancel, outputChTimeout, job)

				if err != nil {
					result.err = err
					running = false
				}
			} else {
				w.logger.Debug("jobs chan closed",
					slog.String("source", "worker.run"),
					slog.String("worker-id", string(w.id)),
				)

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
		w.logger.Debug("output timeout",
			slog.String("source", "Worker.invoke"),
			slog.String("output-channel-timeout", outputChTimeout.String()),
		)

		select {
		case w.outputsChOut <- result:

		case <-parentContext.Done():
			w.logger.Debug("done received",
				slog.String("source", "Worker.invoke"),
				slog.String("worker-id", string(w.id)),
			)

		case <-outputContext.Done():
			w.logger.Debug("cancel - timeout on send",
				slog.String("source", "Worker.invoke"),
				slog.String("worker-id", string(w.id)),
			)

			// ??? err = i18n.NewOutputChTimeoutError()
			err = errors.New("timeout on send")

			parentCancel()
		}
	}

	return err
}
