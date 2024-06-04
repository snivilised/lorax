package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snivilised/lorax/boost"
)

// Demonstrates use of manifold func base worker pool where
// the client manifold func returns an output and an error. An
// output channel is created through which the client receives
// all generated outputs.

const (
	AntsSize        = 1000
	n               = 100000
	OutputChSize    = 10
	Param           = 100
	OutputChTimeout = time.Second / 2 // do not use a value that is similar to interval
	interval        = time.Second / 10
)

func produce(ctx context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Only the producer (observable) knows when the workload is complete
	// but clearly it has no idea when the worker-pool is complete. Initially,
	// one might think that the worker-pool knows when work is complete
	// but this is in correct. The pool only knows when the pool is dormant,
	// not that no more jobs will be submitted.
	// This poses a problem from the perspective of the consumer; it does
	// not know when to exit its output processing loop.
	// What this indicates to us is that the knowledge of end of workload is
	// a combination of multiple events:
	//
	// 1) The producer knows when it will submit no more work
	// 2) The pool knows when all it's workers are dormant
	//
	// A non deterministic way for the consumer to exit it's output processing
	// loop, is to use a timeout. But what is a sensible value? Only the client
	// knows this and even so, it can't really be sure no more outputs will
	// arrive after the timeout; essentially its making an educated guess, which
	// is not reliable.
	//
	for i, n := 0, 100; i < n; i++ {
		_ = pool.Post(ctx, Param)
	}

	pool.EndWork(ctx, interval)
}

func consume(_ context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// We don't need to use a timeout on the observe channel
	// because our producer invokes EndWork, which results in
	// the observe channel being closed, terminating the range.
	// This aspect is specific to this example and clients may
	// have to use different strategies depending on their use-case,
	// eg support for context cancellation.
	//
	for output := range pool.Observe() {
		fmt.Printf("🍒 payload: '%v', id: '%v', seq: '%v' (e: '%v')\n",
			output.Payload, output.ID, output.SequenceNo, output.Error,
		)
	}
}

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := boost.NewManifoldFuncPool(
		ctx, AntsSize, func(input int) (int, error) {
			time.Sleep(time.Duration(input) * time.Millisecond)

			return n + 1, nil
		}, &wg,
		boost.WithOutput(OutputChSize),
	)

	defer pool.Release(ctx)

	if err != nil {
		fmt.Printf("🔥 error creating pool: '%v'\n", err)
		return
	}

	wg.Add(1)
	go produce(ctx, pool, &wg) //nolint:wsl // pendant

	wg.Add(1)
	go consume(ctx, pool, &wg) //nolint:wsl // pendant

	fmt.Printf("pool with func, no of running workers:%d\n",
		pool.Running(),
	)
	wg.Wait()
	fmt.Println("🏁 (manifold-func-pool) FINISHED")
}
