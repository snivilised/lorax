package boost

const (
	MaxWorkers = 100
)

type (
	Job[I any] struct {
		ID         string
		SequenceNo int
		Input      I
	}

	JobOutput[O any] struct {
		ID         string
		SequenceNo int
		Payload    O
		Error      error
	}

	JobStream[I any]  chan Job[I]
	JobStreamR[I any] <-chan Job[I]
	JobStreamW[I any] chan<- Job[I]

	JobOutputStream[O any]  chan JobOutput[O]
	JobOutputStreamR[O any] <-chan JobOutput[O]
	JobOutputStreamW[O any] chan<- JobOutput[O]

	// Duplex represents a channel with multiple views, to be used
	// by clients that need to hand out different ends of the same
	// channel to different entities.
	Duplex[T any] struct {
		Channel  chan T
		ReaderCh <-chan T
		WriterCh chan<- T
	}

	DuplexJobOutput[O any] Duplex[JobOutput[O]]

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

	// Next is a sequential unique id generator func type
	Next func() string
)

type ExecutiveFunc[I, O any] func(j Job[I]) (JobOutput[O], error)

func (f ExecutiveFunc[I, O]) Invoke(j Job[I]) (JobOutput[O], error) {
	return f(j)
}

// NewDuplex creates a new instance of a Duplex with all members populated
func NewDuplex[T any](channel chan T) *Duplex[T] {
	return &Duplex[T]{
		Channel:  channel,
		ReaderCh: channel,
		WriterCh: channel,
	}
}
