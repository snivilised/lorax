package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/snivilised/lorax/async"
)

type ProviderFunc[I any] func() I

type Producer[I, R any] struct {
	sequenceNo  int
	JobsCh      async.JobStream[I]
	quit        *sync.WaitGroup
	Count       int
	provider    ProviderFunc[I]
	delay       int
	terminateCh chan string
}

// The producer owns the Jobs channel as it knows when to close it. This producer is
// a fake producer and exposes a stop method that the client go routing can call to
// indicate end of the work load.
func StartProducer[I, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	capacity int,
	provider ProviderFunc[I],
	delay int,
) *Producer[I, R] {
	if delay == 0 {
		panic(fmt.Sprintf("Invalid delay requested: '%v'", delay))
	}

	producer := Producer[I, R]{
		JobsCh:      make(async.JobStream[I], capacity),
		quit:        wg,
		provider:    provider,
		delay:       delay,
		terminateCh: make(chan string),
	}
	go producer.run(ctx)

	return &producer
}

func (p *Producer[I, R]) run(ctx context.Context) {
	defer func() {
		close(p.JobsCh)
		p.quit.Done()
		fmt.Printf(">>>> producer.run - finished (QUIT). ✨✨✨ \n")
	}()

	fmt.Printf(">>>> ✨ producer.run ...(ctx:%+v)\n", ctx)

	// ⚠️ The producer is not responding to the done message, and therefore
	// doesn't get to exit its loop and then invoke p.quit.Done(), hence
	// the deadlock.
	//
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false

			fmt.Println(">>>> 💠 producer.run - done received ⛔⛔⛔")

		case <-p.terminateCh:
			running = false
			fmt.Printf(">>>> ✨ producer.run - termination detected (running: %v)\n", running)

		case <-time.After(time.Second / time.Duration(p.delay)):
			fmt.Printf(">>>> ✨ producer.run - default (running: %v) ...\n", running)

			if !p.item(ctx) {
				running = false
			}
		}
	}
}

func (p *Producer[I, R]) item(ctx context.Context) bool {
	result := true
	i := p.provider()
	j := async.Job[I]{
		ID:    fmt.Sprintf("JOB-ID:%v", uuid.NewString()),
		Input: i,
	}

	fmt.Printf(">>>> ✨ producer.item, 🟠 waiting to post item: '%+v'\n", i)

	select {
	case <-ctx.Done():
		fmt.Println(">>>> 💠 producer.item - done received ⛔⛔⛔")

		result = false

	case p.JobsCh <- j:
	}
	p.Count++

	if result {
		fmt.Printf(">>>> ✨ producer.item, 🟢 posted item: '%+v'\n", i)
	} else {
		fmt.Printf(">>>> ✨ producer.item, 🔴 item NOT posted: '%+v'\n", i)
	}

	return result
}

func (p *Producer[I, R]) Stop() {
	fmt.Println(">>>> 🧲 producer terminating ...")
	p.terminateCh <- "done"
	close(p.terminateCh)
}

// StopProducerAfter, run in a new go routine
func StopProducerAfter[I, R any](
	ctx context.Context,
	producer *Producer[I, R],
	delay time.Duration,
) {
	fmt.Printf("		>>> 💤 StopAfter - Sleeping before requesting stop (%v) ...\n", delay)
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}

	producer.Stop()
	fmt.Printf("		>>> StopAfter - 🍧🍧🍧 stop submitted.\n")
}

func CancelProducerAfter[I, R any](
	_ context.Context,
	delay time.Duration,
	cancellation ...context.CancelFunc,
) {
	fmt.Printf("		>>> 💤 CancelAfter - Sleeping before requesting cancellation (%v) ...\n", delay)
	<-time.After(delay)

	// we should always expect to get a cancel function back, even if we don't
	// ever use it, so it is still relevant to get it in the stop test case
	//
	if len(cancellation) > 0 {
		cancel := cancellation[0]

		fmt.Printf("		>>> CancelAfter - 🛑🛑🛑 cancellation submitted.\n")
		cancel()
		fmt.Printf("		>>> CancelAfter - ➖➖➖ CANCELLED\n")
	} else {
		fmt.Printf("		>>> CancelAfter(noc) - ✖️✖️✖️ cancellation attempt benign.\n")
	}
}
