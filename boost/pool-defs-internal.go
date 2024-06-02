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

	workerWrapperL[I any, O any] struct {
		core *workerL[I, O]
	}

	workersCollectionL[I, O any] map[workerID]*workerWrapperL[I, O]
)

// Worker pool types:
//
// 🍺 ManifoldFuncPool (to be used by traverse):
// description: this is the most comprehensive pool type with return
// semantics. It is functional meaning that the pool is defined by a
// predefined executive function.
// ants: PoolWithFunc
// post(ants): Invoke
// job(Param): Job(I)
// job-return: JobOutput(O), error
// job-input-stream(client-side): JobStreamW[I]
// job-input-stream(pool-side): JobStreamR[I]
// returns err: true
// observable: JobOutputStreamR(O)
// start: returns observable stream, completion stream
// pool-result: tbd (this is the result that represents the overall pool result.
// If pool shuts down as a result of premature error or ctrl-c abort, then this
// will be reflected in the pool's result).
//
// 🍺 ManifoldTaskPool:
// description: like ManifoldFuncPool but accepts task based jobs meaning each
// job can be any function as opposed to be being a pre-defined function registered
// with the pool. Each job accepts an input I and emits an output O with an error.
// ants: Pool
// post(ants): Submit
// job(Param): Job(func(I) JobOutput(O), error)
// job-return: JobOutput(O), error
// job-input-stream(client-side): JobStreamW[I]
// job-input-stream(pool-side): JobStreamR[I]
// returns err: true
// observable: JobOutputStreamR(O)
// start: returns observable stream, completion stream
// pool-result: yes
//
// 🍺 FuncPoolE
// description: A simple functional pool with fire and return semantics. Client
// submits jobs with only an error return value.
// ants: PoolWithFunc
// post(ants): Invoke
// job(Param): Job(I)
// job-return: none; error only
// job-input-stream(client-side): JobStreamW[I]
// job-input-stream(pool-side): JobStreamR[I]
// returns err: yes
// observable: none
// start: returns completion stream
// pool-result: yes
//
// 🍺 FuncPool
// description: A simple functional pool with fire and forget semantics. Client
// submits jobs with no return value
// ants: PoolWithFunc
// post(ants): Invoke
// job(Param): Job(I)
// job-return: none
// job-input-stream(client-side): JobStreamW[I]
// job-input-stream(pool-side): JobStreamR[I]
// returns err: no
// observable: none
// start: returns completion stream
// pool-result: yes
//
// 🍺 TaskPoolE
// description: accepts task based jobs. Each job accepts an input I and
// emits only an error return value.
// ants: Pool
// post(ants): Submit
// job(Param): Job(func(I) error)
// job-return: error
// job-input-stream(client-side): JobStreamW[I]
// job-input-stream(pool-side): JobStreamR[I]
// returns err: true
// observable: JobOutputStreamR(O)
// start: returns observable stream, completion stream
// pool-result: yes
//
