package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/snivilised/lorax/boost"
)

type Consumer[O any] struct {
	quitter     boost.AnnotatedWgQuitter
	RoutineName boost.GoRoutineName
	interval    time.Duration
	OutputsChIn boost.JobOutputStreamR[O]
	Count       int
}

func StartConsumer[O any](
	parentContext context.Context,
	quitter boost.AnnotatedWgQuitter,
	outputsChIn boost.JobOutputStreamR[O],
	interval time.Duration,
) *Consumer[O] {
	consumer := &Consumer[O]{
		quitter:     quitter,
		RoutineName: boost.GoRoutineName("💠 consumer"),
		interval:    interval,
		OutputsChIn: outputsChIn,
	}

	go consumer.run(parentContext)

	return consumer
}

func (c *Consumer[O]) run(parentContext context.Context) {
	defer func() {
		c.quitter.Done(c.RoutineName)
		fmt.Printf("<<<< 💠 consumer.run - finished (QUIT). 💠💠💠 \n")
	}()
	fmt.Printf("<<<< 💠 consumer.run ...(ctx:%+v)\n", parentContext)

	for running := true; running; {
		<-time.After(c.interval)
		select {
		case <-parentContext.Done():
			running = false

			fmt.Println("<<<< 💠 consumer.run - done received 💔💔💔")

		case result, ok := <-c.OutputsChIn:
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
