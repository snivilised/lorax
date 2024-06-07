# 🔆 lorax/boost: ___🐜🐜🐜 ants based worker pool___

<!-- MD013/Line Length -->
<!-- MarkDownLint-disable MD013 -->

<!-- MD014/commands-show-output: Dollar signs used before commands without showing output mark down lint -->
<!-- MarkDownLint-disable MD014 -->

<!-- MD033/no-inline-html: Inline HTML -->
<!-- MarkDownLint-disable MD033 -->

<!-- MD040/fenced-code-language: Fenced code blocks should have a language specified -->
<!-- MarkDownLint-disable MD040 -->

<!-- MD028/no-blanks-blockquote: Blank line inside blockquote -->
<!-- MarkDownLint-disable MD028 -->

<!-- MD010/no-hard-tabs: Hard tabs -->
<!-- MarkDownLint-disable MD010 -->

## 📚 Usage

Please refer to the documentation already available at [🐜🐜🐜 ___ants___](https://github.com/panjf2000/ants). The documentation here focuses on the functionality provided that augments the underlying ___ants___ implementation.

Included in this repo are some executable examples that help explain how the pool works and demonstrates some key characteristics that will aid in understanding of how to correctly use this package. For more detailed explanation of the _Options_, the reader is encouraged to read the ants documentation.

## 🎀 Additional Features

The ___ants___ implementation was chosen because it has already proven itself in production, having a wide install base and addresses scalability and reliability issues. However after review of its features, it was discovered that there were a few supplementary features that it did not possess including the following:

+ __no top level client defined context__: this means there is no way for the client to cancel an operation using idiomatic ___Go___ techniques.
+ __no job return error__: that is to say, whenever a job is executed, there is no notification of wether it executed successfully or not. Rather, it has been implemented on a _fire and forget_ basis.
+ __no job output__: similar to the lack of an error result for each job, there is no way for the result of an operation to be collated; eg the client may request that the pool perform some task that contains a result. In the ants implementation, there is no native way to return an output for each job.
+ __no input channel__: the client needs direct access to the pool instance in order to submit tasks with a function call. However, there are benefits including but not limited to reduced coupling. With an input channel, the client can pass this channel to another entity capable of generating a workload without having direct access to the pool itself, all they need to to do is simply write to the channel.

## 💫 ManifoldFuncPool

### 🚀 Quick start

#### 📌 Create pool with output

```go
	pool, err := boost.NewManifoldFuncPool(
		ctx, func(input int) (int, error) {
			// client implementation; output = something

			return output, nil
		}, &wg,
		boost.WithSize(PoolSize),
		boost.WithOutput(OutputChSize, CheckCloseInterval, TimeoutOnSend),
	)
```

Creates an _int_ based manifold worker pool. The ___ManifoldFuncPool___ is a generic whose type parameters represents the Input type _I_ and the output type _O_. In this example, the input and output types are both _int_ as denoted by the signature of the manifold function:

> func(input int) (int, error)

NB: It is not mandatory to require workers to send outputs. If the ___WithOutput___ option is not specified, then an output will still occur, but will be ignored.

#### 📌 Submit work

There are 2 ways to submit work to the pool, either directly or by input channel

+ direct(Post):

```go
  pool.Post(ctx, 42)

  ...
  pool.Conclude(ctx)
```

Sends a job to the pool with int based input value 42. Typically, the Post would be issued multiple times as needs demands. At some point we are done submitting work. The end of the workload needs to be communicated to the pool. This is the purpose of invoking Conclude.

+ via input channel(Source):

```go
  inputCh := pool.Source(ctx, wg)
  inputCh <- 42

  ...
  close(inputCh)
```

Sends a job to the pool with int based input value 42, via the input channel. At the end of the workload, all we need to do is close the channel; we do not need to invoke ___Conclude___ explicitly as this is done automatically on our behalf as a result of the channel closure.

#### 📌 Consume outputs

Outputs can be consumed simply by invoking ___pool.Observe___ which returns a channel:

```go
  select {
  case output := <-pool.Observe():
    fmt.Printf("🍒 payload: '%v', id: '%v', seq: '%v' (e: '%v')\n",
      output.Payload, output.ID, output.SequenceNo, output.Error, 
    )
  case <-ctx.Done():
    return
  }
```

Each output is represented by a ___JobOutput___ which contains a _Payload_ field representing the job's result and some supplementary meta data fields, including a sequence number and a job ID.

It is possible to range over the output channel as illustrated:

```go
	for output := range pool.Observe() {
		fmt.Printf("🍒 payload: '%v', id: '%v', seq: '%v' (e: '%v')\n",
			output.Payload, output.ID, output.SequenceNo, output.Error,
		)
	}
```

This will work in success cases, but what happens if a worker send timeout occurs? The worker will send a cancellation request and the context will be cancelled as a result. But since the range operator is not pre-empted as a result of this cancellation, it will continue to block, waiting for either more content or channel closure. If the main Go routine is blocking on a WaitGroup, which it almost certainly should be, the program will deadlock on the wait. For this reason, it is recommended to use a select statement as shown.

#### 📌 Monitor the cancellation channel

Currently, the only reason for a worker to request a cancellation is that it is unable to send an output. Any request cancellation must be addressed by the client, this means invoking the cancel function associated with the context.

The client can delegate this responsibility to a pre defined function in boost: ___StartCancellationMonitor___:

```go
	if cc := pool.CancelCh(); cc != nil {
		boost.StartCancellationMonitor(ctx, cancel, &wg, cc, func() {
			fmt.Print("🔴 cancellation received, cancelling...\n")
		})
	}
```

Note, the client is able to pass in a callback function which is invoked, if cancellation occurs. Also, note that there is no need to increment the wait group as that is done internally.

## 📝 Design

In designing the augmented functionality, it was discovered that there could conceivably be more than 1 abstraction, depending on the client's needs. From the perspective of ___snivilised___ projects, the key requirement was to have a pool that could execute jobs and for each one, return an error code and an output. The name given to this implementation is the ___ManifoldFuncPool___.

In ___ants___, there are 2 main implementations of worker pool, ___Pool___ or ___PoolFunc___.

+ ___Pool___: accepts new jobs represented by a function. Each function can implement any logic, so the pool is in fact able to execute a stream of heterogenous tasks.
+ ___PoolFunc___: the pool is created with a pre-defined function and accepts new jobs specified as an input to this pool function. So every job the pool executes, runs the same functionality but with a different input.

___ManifoldFuncPool___ is based on the ___PoolFunc___ implementation. However, ___PoolFunc___ does not return either an output or an error, ___ManifoldFuncPool___ allows for this behaviour by allowing the client to define a function (_manifold function_) whose signature allows for an input of a specific type, along with an output and error. ___ManifoldFuncPool___ therefore provides a mapping from the _manifold function_ to the ants function (_PoolFunc_).

As previously mentioned, ___boost___ could provide many more worker pool abstractions, eg there could be a ___ManifoldTaskPool___ based upon the ___Pool___ implementation. However, ___ManifoldTaskPool___ is not currently defined as there is no established need for one. Similarly, boost could provide a ___PoolFunc___ based pool whose client function only returns an error. Future versions of lorax/boost could provide these alternative implementations if such a need arises.

### Context

The ___NewManifoldFuncPool___ constructor function accepts a context, that works in exactly the way one would expect. Any internal Go routine works with this context. If the client cancels this context, then this will be propagated to all child Go routines including the workers in the pool.

### Cancellation

The need to send output back to the client for each job presents us with an additional problem. Once the need for output has been declared via use of the ___WithOutput___ option, there is an obligation on the client to consume it. Failure to consume, will result in the eventual blockage of the entire worker pool; the pool will get to a state where all workers are blocking on their attempt to send the output, the output buffer is full and new incoming requests can no longer be dispatched to workers, as they are all busy, resulting in deadlock. This may just be a programming error, but it would be undesirable for the pool to simply end up in deadlock.

This has been alleviated by the use of a timeout mechanism. The ___WithOutput___ option takes a timeout parameter defined as a ___time.Duration___. When the worker timeouts out attempting to send the output, it will then send a cancellation request back to the client via a separate cancellation channel (obtained by invoking ___ManifoldFuncPool.CancelCh___).

Since context cancellation should only be initiated by the client, the onus is on them to cancel the context. However, the way in which this would be done amounts to some boilerplate code, so ___boost___ also provides this as a function ___StartCancellationMonitor___, which starts a Go routine that monitors the cancellation channel for requests and on seeing one, cancels the associated context. This results in all child Go routines abandoning their work when they are able and exiting gracefully. This means that we can avoid the deadlock and leaked Go routines.

### Conclude

The pool needs to close the output channel so the consumer knows to exit it's read loop, but it can only do so once its clear there are no more outstanding jobs to complete and all workers are idle. We can't close the channel prematurely as that would result in a panic when a worker attempts to send the output.

Typically, the use of the ___WithOutput___ operator looks like this:

> boost.WithOutput(OutputChSize, CheckCloseInterval, TimeoutOnSend)

The ___CheckCloseInterval___ parameter is internally required by ___pool.Conclude___. To counter the problem described above, ___Conclude___ needs to check if its safe to close the output channel, periodically, which is implemented within another Go routine. ___CheckCloseInterval___ denotes the amount of time it will wait before checking again.
