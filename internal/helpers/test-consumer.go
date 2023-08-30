package helpers

import (
	"context"
	"fmt"

	"github.com/snivilised/lorax/boost"
)

type Consumer[O any] struct {
	quitter     boost.AnnotatedWgQuitter
	RoutineName boost.GoRoutineName
	OutputsChIn boost.OutputStreamR[O]
	Count       int
}

func StartConsumer[O any](
	ctx context.Context,
	quitter boost.AnnotatedWgQuitter,
	outputsChIn boost.OutputStreamR[O],
) *Consumer[O] {
	consumer := &Consumer[O]{
		quitter:     quitter,
		RoutineName: boost.GoRoutineName("💠 consumer"),
		OutputsChIn: outputsChIn,
	}
	go consumer.run(ctx)

	return consumer
}

func (c *Consumer[O]) run(ctx context.Context) {
	defer func() {
		c.quitter.Done(c.RoutineName)
		fmt.Printf("<<<< 💠 consumer.run - finished (QUIT). 💠💠💠 \n")
	}()
	fmt.Printf("<<<< 💠 consumer.run ...(ctx:%+v)\n", ctx)

	for running := true; running; {
		select {
		case <-ctx.Done():
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
