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
	ID         string
	Input      I
	SequenceNo int
}

type ExecutiveFunc[I, O any] func(j Job[I]) (JobOutput[O], error)

func (f ExecutiveFunc[I, O]) Invoke(j Job[I]) (JobOutput[O], error) {
	return f(j)
}

type JobOutput[O any] struct {
	Payload O
}

type JobStream[I any] chan Job[I]
type JobStreamR[I any] <-chan Job[I]
type JobStreamW[I any] chan<- Job[I]

type OutputStream[O any] chan JobOutput[O]
type OutputStreamR[O any] <-chan JobOutput[O]
type OutputStreamW[O any] chan<- JobOutput[O]

type CancelWorkSignal struct{}
type CancelStream = chan CancelWorkSignal
type CancelStreamR = <-chan CancelWorkSignal
type CancelStreamW = chan<- CancelWorkSignal

type WorkerID string
type FinishedStream = chan WorkerID
type FinishedStreamR = <-chan WorkerID
type FinishedStreamW = chan<- WorkerID
