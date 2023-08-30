package boost

const (
	MaxWorkers = 100
)

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
