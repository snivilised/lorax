package async

const (
	// TODO: This is just temporary, channel size definition still needs to be
	// fine tuned
	//
	DefaultChSize = 100
)

type workerWrapper[I any, R any] struct {
	cancelChOut chan<- CancelWorkSignal
	core        *worker[I, R]
}

type workersCollection[I, R any] map[WorkerID]*workerWrapper[I, R]
