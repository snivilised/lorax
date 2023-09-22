package boost

const (
	// TODO: This is just temporary, channel size definition still needs to be
	// fine tuned
	//
	DefaultChSize = 100
)

type (
	workerID             string
	workerFinishedResult struct {
		id  workerID
		err error
	}

	finishedStream  = chan *workerFinishedResult
	finishedStreamR = <-chan *workerFinishedResult
	finishedStreamW = chan<- *workerFinishedResult

	workerWrapper[I any, O any] struct {
		core *worker[I, O]
	}

	workersCollection[I, O any] map[workerID]*workerWrapper[I, O]
)
