package async

const (
	// TODO: This is just temporary, channel size definition still needs to be
	// fine tuned
	//
	DefaultChSize = 100
)

type workerWrapper[I any, O any] struct {
	cancelChOut chan<- CancelWorkSignal
	core        *worker[I, O]
}

type workersCollection[I, O any] map[WorkerID]*workerWrapper[I, O]
