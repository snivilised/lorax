package async

const (
	MaxWorkers = 100
)

// Job, this definition is very rudimentary and bears no resemblance to the final
// version. The job definition should be data driven not functionally driven. We
// could have a bind function/method that would bind data to the job fn.
//
// Job also needs a sequence number (can't be defined yet because Job is just a function,
// and there does not allow for meta data). What we could do is to use a functional
// composition technique that allows us to create compound functionality. Will need to
// refresh knowledge of functional programming, see ramda.
//

type Job[I any] struct {
	ID    string
	Input I
}

type ExecutiveFunc[I, R any] func(j Job[I]) (JobResult[R], error)

func (f ExecutiveFunc[I, R]) Invoke(j Job[I]) (JobResult[R], error) {
	return f(j)
}

type JobResult[R any] struct {
	Payload R
}

type JobStream[I any] chan Job[I]
type JobStreamIn[I any] <-chan Job[I]
type JobStreamOut[I any] chan<- Job[I]

type ResultStream[R any] chan JobResult[R]
type ResultStreamIn[R any] <-chan JobResult[R]
type ResultStreamOut[R any] chan<- JobResult[R]

type CancelWorkSignal struct{}
type CancelStream = chan CancelWorkSignal
type CancelStreamIn = <-chan CancelWorkSignal
type CancelStreamOut = chan<- CancelWorkSignal

type WorkerID string
type FinishedStream = chan WorkerID
type FinishedStreamIn = <-chan WorkerID
type FinishedStreamOut = chan<- WorkerID
