package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/snivilised/lorax/async"
)

const (
	BatchSize = 10
)

type ProviderFn[I any] func() I

type Producer[I, R any] struct {
	sequenceNo  int
	audience    []string
	batchSize   int
	JobsCh      async.JobStream[I]
	quit        *sync.WaitGroup
	Count       int
	provider    ProviderFn[I]
	delay       int
	terminateCh chan string
}

// The producer owns the Jobs channel as it knows when to close it. This producer is
// a fake producer and exposes a stop method that the client go routing can call to
// indicate end of the work load.
func NewProducer[I, R any](ctx context.Context,
	wg *sync.WaitGroup,
	capacity int,
	provider ProviderFn[I],
	delay int,
) *Producer[I, R] {
	if delay == 0 {
		panic(fmt.Sprintf("Invalid delay requested: '%v'", delay))
	}

	producer := Producer[I, R]{
		audience: []string{ // REDUNDANT, done by client via the provider function
			"paul", "phil", "lindsey", "kaz", "kerry",
			"nick", "john", "raj", "jim", "mark", "robyn",
		},
		batchSize:   BatchSize,
		JobsCh:      make(async.JobStream[I], capacity),
		quit:        wg,
		provider:    provider,
		delay:       delay,
		terminateCh: make(chan string, async.DefaultChSize),
	}
	go producer.start(ctx)

	return &producer
}

func (p *Producer[I, R]) start(ctx context.Context) {
	defer func() {
		close(p.JobsCh)
		fmt.Printf("===> producer finished (Quit). ✨✨✨ \n")
		p.quit.Done()
	}()

	fmt.Printf("===> ✨ producer.start ...\n")

	for running := true; running; {
		select {
		case <-ctx.Done():
			fmt.Println("---> 💠 producer.start - done received ⛔⛔⛔")

			running = false

		case <-p.terminateCh:
			running = false
			fmt.Printf("---> ✨ producer termination detected (running: %v)\n", running)

		default:
			fmt.Printf("---> ✨ producer.start/default(running: %v) ...\n", running)
			p.item()

			time.Sleep(time.Second / time.Duration(p.delay))
		}
	}
}

func (p *Producer[I, R]) item() {
	i := p.provider()
	j := async.Job[I]{
		Input: i,
	}
	p.JobsCh <- j
	p.Count++

	fmt.Printf("===> ✨ producer.item, posted item: '%+v'\n", i)
}

func (p *Producer[I, R]) Stop() {
	fmt.Println("---> 🧲 terminating ...")
	p.terminateCh <- "done"
	close(p.terminateCh)
}
