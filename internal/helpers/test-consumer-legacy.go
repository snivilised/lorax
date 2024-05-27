package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/snivilised/lorax/boost"
)

type ConsumerL[O any] struct {
	quitter     boost.AnnotatedWgQuitter
	RoutineName boost.GoRoutineName
	interval    time.Duration
	OutputsChIn boost.JobOutputStreamR[O]
	Count       int
	verbose     bool
}

func StartConsumerL[O any](
	parentContext context.Context,
	quitter boost.AnnotatedWgQuitter,
	outputsChIn boost.JobOutputStreamR[O],
	interval time.Duration,
	verbose bool,
) *ConsumerL[O] {
	consumer := &ConsumerL[O]{
		quitter:     quitter,
		RoutineName: boost.GoRoutineName("💠 consumer"),
		interval:    interval,
		OutputsChIn: outputsChIn,
		verbose:     verbose,
	}

	go consumer.run(parentContext)

	return consumer
}

func (c *ConsumerL[O]) run(parentContext context.Context) {
	defer func() {
		c.quitter.Done(c.RoutineName)
		if c.verbose {
			fmt.Printf("<<<< 💠 consumer.run - finished (QUIT). 💠💠💠 \n")
		}
	}()
	if c.verbose {
		fmt.Printf("<<<< 💠 consumer.run ...(ctx:%+v)\n", parentContext)
	}

	for running := true; running; {
		<-time.After(c.interval)
		select {
		case <-parentContext.Done():
			running = false

			if c.verbose {
				fmt.Println("<<<< 💠 consumer.run - done received 💔💔💔")
			}

		case result, ok := <-c.OutputsChIn:
			if ok {
				c.Count++
				if c.verbose {
					fmt.Printf("<<<< 💠 consumer.run - new result arrived(#%v): '%+v' \n",
						c.Count, result.Payload,
					)
				}
			} else {
				running = false
				if c.verbose {
					fmt.Printf("<<<< 💠 consumer.run - no more results available (running: %+v)\n", running)
				}
			}
		}
	}
}
