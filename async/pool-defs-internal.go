package async

type workerInfo[I, R any] struct {
	job         Job[I]
	resultsOut  ResultStreamOut[R]
	finishedOut FinishedStreamOut
}

const (
	// TODO: This is just temporary, channel size definition still needs to be
	// fine tuned
	//
	DefaultChSize = 100
)
