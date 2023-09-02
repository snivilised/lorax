package boost

const (
	MaxWorkers = 100
)

type (
	Job[I any] struct {
		ID         string
		Input      I
		SequenceNo int
	}

	JobOutput[O any] struct {
		Payload O
	}

	JobStream[I any]  chan Job[I]
	JobStreamR[I any] <-chan Job[I]
	JobStreamW[I any] chan<- Job[I]

	OutputStream[O any]  chan JobOutput[O]
	OutputStreamR[O any] <-chan JobOutput[O]
	OutputStreamW[O any] chan<- JobOutput[O]

	CancelWorkSignal struct{}
	CancelStream     = chan CancelWorkSignal
	CancelStreamR    = <-chan CancelWorkSignal
	CancelStreamW    = chan<- CancelWorkSignal

	PoolResult struct {
		Error error
	}

	PoolResultStream  = chan *PoolResult
	PoolResultStreamR = <-chan *PoolResult
	PoolResultStreamW = chan<- *PoolResult
)

type ExecutiveFunc[I, O any] func(j Job[I]) (JobOutput[O], error)

func (f ExecutiveFunc[I, O]) Invoke(j Job[I]) (JobOutput[O], error) {
	return f(j)
}
