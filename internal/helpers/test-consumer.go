package helpers

import (
	"context"
	"fmt"
	"sync"

	"github.com/snivilised/lorax/async"
)

type Consumer[R any] struct {
	ResultsCh <-chan async.JobResult[R]
	quit      *sync.WaitGroup
	Count     int
}

func StartConsumer[R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	resultsCh <-chan async.JobResult[R],
) *Consumer[R] {
	consumer := &Consumer[R]{
		ResultsCh: resultsCh,
		quit:      wg,
	}
	go consumer.run(ctx)

	return consumer
}

func (c *Consumer[R]) run(ctx context.Context) {
	defer func() {
		c.quit.Done()
		fmt.Printf("<<<< consumer.run - finished (QUIT). 💠💠💠 \n")
	}()
	fmt.Printf("<<<< 💠 consumer.run ...\n")

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false

			fmt.Println("<<<< 💠 consumer.run - done received 💔💔💔")

		case result, ok := <-c.ResultsCh:
			if ok {
				c.Count++
				fmt.Printf("<<<< 💠 consumer.run - new result arrived(#%v): '%+v' \n",
					c.Count, result.Payload,
				)
			} else {
				running = false
				fmt.Printf("<<<< 💠 consumer.run - no more results available (running: %+v)\n", running)
			}
		}
	}
}
