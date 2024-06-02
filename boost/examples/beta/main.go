package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snivilised/lorax/boost"
)

// Demonstrates that when all workers are engaged and the pool is at capacity,
// new incoming jobs are blocked, until a worker becomes free. The invoked function
// takes a second to complete. The PRE and POST indicators reflect this:
//
// PRE: <--- (n: 0) [13:41:49] 🍋
// POST: <--- (n: 0) [13:41:49] 🍊
// PRE: <--- (n: 1) [13:41:49] 🍋
// POST: <--- (n: 1) [13:41:49] 🍊
// PRE: <--- (n: 2) [13:41:49] 🍋
// POST: <--- (n: 2) [13:41:49] 🍊
// PRE: <--- (n: 3) [13:41:49] 🍋
// => running: '3')
// <--- (n: 2)🍒
// => running: '3')
// <--- (n: 1)🍒
// => running: '3')
// <--- (n: 0)🍒
// POST: <--- (n: 3) [13:41:50] 🍊
//
// Considering the above, whilst the pool is not at capacity, each new submission is
// executed immediately, as a new worker can be allocated to those jobs (n=0..2).
// Once the pool has reached capacity (n=3), the PRE is blocked, because its corresponding
// POST doesn't happen until a second later; this illustrates the blocking.
//

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const NoW = 3

	pool, _ := boost.NewTaskPool[int, int](ctx, NoW, &wg)

	defer pool.Release(ctx)

	for i := 0; i < 30; i++ { // producer
		fmt.Printf("PRE: <--- (n: %v) [%v] 🍋 \n", i, time.Now().Format(time.TimeOnly))
		_ = pool.Post(ctx, func() {
			fmt.Printf("=> running: '%v')\n", pool.Running())
			fmt.Printf("<--- (n: %v)🍒 \n", i)
			time.Sleep(time.Second)
		})
		fmt.Printf("POST: <--- (n: %v) [%v] 🍊 \n", i, time.Now().Format(time.TimeOnly))
	}

	fmt.Printf("task pool, running workers number:%d\n",
		pool.Running(),
	)
	wg.Wait()
	fmt.Println("🏁 (task-pool) FINISHED")
}
