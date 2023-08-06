package helpers

import (
	"context"
	"fmt"
	"sync"
)

type Consumer[R any] struct {
	ResultsCh <-chan R
	quit      *sync.WaitGroup
	Count     int
}

func NewConsumer[R any](ctx context.Context, wg *sync.WaitGroup, resultsCh <-chan R) *Consumer[R] {
	consumer := &Consumer[R]{
		ResultsCh: resultsCh,
		quit:      wg,
	}
	go consumer.start(ctx)

	return consumer
}

func (c *Consumer[R]) start(ctx context.Context) {
	defer func() {
		fmt.Printf("===> consumer finished (Quit). 💠💠💠 \n")
		c.quit.Done()
	}()
	fmt.Printf("===> 💠 consumer.start ...\n")

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Println("---> 💠 consumer.start - done received 💔💔💔")

			running = false

		case result, ok := <-c.ResultsCh:
			if ok {
				c.Count++
				fmt.Printf("---> 💠 consumer.start new result arrived(#%v): '%+v' \n", c.Count, result)
			} else {
				running = false
				fmt.Printf("---> 💠 consumer.start no more results available (running: %+v)\n", running)
			}
		}
	}
}
