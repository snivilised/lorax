package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snivilised/lorax/boost"
)

type termination string
type terminationDuplex *boost.Duplex[termination]

type ProviderFuncL[I any] func() I

type ProducerL[I, O any] struct {
	quitter      boost.AnnotatedWgQuitter
	RoutineName  boost.GoRoutineName
	sequenceNo   int
	provider     ProviderFuncL[I]
	interval     time.Duration
	terminateDup terminationDuplex
	JobsCh       boost.JobStream[I]
	Count        int
	verbose      bool
}

// The producer owns the Jobs channel as it knows when to close it. This producer is
// a fake producer and exposes a stop method that the client go routine can call to
// indicate end of the work load.
func StartProducerL[I, O any](
	parentContext context.Context,
	quitter boost.AnnotatedWgQuitter,
	capacity int,
	provider ProviderFuncL[I],
	interval time.Duration,
	verbose bool,
) *ProducerL[I, O] {
	if interval == 0 {
		panic(fmt.Sprintf("Invalid delay requested: '%v'", interval))
	}

	producer := ProducerL[I, O]{
		quitter:      quitter,
		RoutineName:  boost.GoRoutineName("✨ producer"),
		provider:     provider,
		interval:     interval,
		terminateDup: boost.NewDuplex(make(chan termination)),
		JobsCh:       make(boost.JobStream[I], capacity),
		verbose:      verbose,
	}

	go producer.run(parentContext)

	return &producer
}

func (p *ProducerL[I, O]) run(parentContext context.Context) {
	defer func() {
		close(p.JobsCh)
		p.quitter.Done(p.RoutineName)

		if p.verbose {
			fmt.Printf(">>>> ✨ producer.run - finished (QUIT). ✨✨✨ \n")
		}
	}()
	if p.verbose {
		fmt.Printf(">>>> ✨ producer.run ...(ctx:%+v)\n", parentContext)
	}

	for running := true; running; {
		select {
		case <-parentContext.Done():
			running = false

			if p.verbose {
				fmt.Println(">>>> ✨ producer.run - done received ⛔⛔⛔")
			}

		case <-p.terminateDup.ReaderCh:
			running = false

			if p.verbose {
				fmt.Printf(">>>> ✨ producer.run - termination detected (running: %v)\n", running)
			}

		case <-time.After(p.interval):
			if p.verbose {
				fmt.Printf(">>>> ✨ producer.run - default (running: %v) ...\n", running)
			}

			if !p.item(parentContext) {
				running = false
			}
		}
	}
}

func (p *ProducerL[I, O]) item(parentContext context.Context) bool {
	p.sequenceNo++
	p.Count++

	result := true
	i := p.provider()
	j := boost.Job[I]{
		ID:         fmt.Sprintf("JOB-ID:%v", uuid.NewString()),
		Input:      i,
		SequenceNo: p.sequenceNo,
	}

	if p.verbose {
		fmt.Printf(">>>> ✨ producer.item, 🟠 waiting to post item: '%+v'\n", i)
	}

	select {
	case <-parentContext.Done():
		if p.verbose {
			fmt.Println(">>>> ✨ producer.item - done received ⛔⛔⛔")
		}

		result = false

	case p.JobsCh <- j:
	}

	if p.verbose {
		if result {
			fmt.Printf(">>>> ✨ producer.item, 🟢 posted item: '%+v'\n", i)
		} else {
			fmt.Printf(">>>> ✨ producer.item, 🔴 item NOT posted: '%+v'\n", i)
		}
	}

	return result
}

func (p *ProducerL[I, O]) Stop() {
	if p.verbose {
		fmt.Println(">>>> 🧲 producer terminating ...")
	}

	p.terminateDup.WriterCh <- termination("done")
	close(p.terminateDup.Channel)
}

// StopProducerAfter, run in a new go routine
func StopProducerAfter[I, O any](
	parentContext context.Context,
	producer *ProducerL[I, O],
	delay time.Duration,
	verbose bool,
) {
	if verbose {
		fmt.Printf("		>>> 💤 StopAfter - Sleeping before requesting stop (%v) ...\n", delay)
	}

	select {
	case <-parentContext.Done():
	case <-time.After(delay):
	}

	producer.Stop()

	if verbose {
		fmt.Printf("		>>> StopAfter - 🍧🍧🍧 stop submitted.\n")
	}
}

func CancelProducerAfterL[I, O any](
	delay time.Duration,
	parentCancel context.CancelFunc,
	verbose bool,
) {
	if verbose {
		fmt.Printf("		>>> 💤 CancelAfter - Sleeping before requesting cancellation (%v) ...\n", delay)
	}
	<-time.After(delay)

	// we should always expect to get a cancel function back, even if we don't
	// ever use it, so it is still relevant to get it in the stop test case
	//
	if parentCancel != nil {
		if verbose {
			fmt.Printf("		>>> CancelAfter - 🛑🛑🛑 cancellation submitted.\n")
		}
		parentCancel()

		if verbose {
			fmt.Printf("		>>> CancelAfter - ➖➖➖ CANCELLED\n")
		}
	} else if verbose {
		fmt.Printf("		>>> CancelAfter(noc) - ✖️✖️✖️ cancellation attempt benign.\n")
	}
}
