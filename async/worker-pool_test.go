package async_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/lorax/async"
	"github.com/snivilised/lorax/internal/helpers"
)

const DefaultNoWorkers = 5

func init() { rand.Seed(time.Now().Unix()) }

// TerminatorFunc brings the work pool processing to an end, eg
// by stopping or cancellation after the requested amount of time.
type TerminatorFunc[I, R any] func(ctx context.Context, delay time.Duration, funcs ...context.CancelFunc)

func (f TerminatorFunc[I, R]) After(ctx context.Context, delay time.Duration, funcs ...context.CancelFunc) {
	f(ctx, delay, funcs...)
}

const (
	JobChSize    = 10
	ResultChSize = 10
	Delay        = 750
)

var audience = []string{
	"👻 caspar",
	"🧙 gandalf",
	"😺 garfield",
	"👺 gobby",
	"👿 nick",
	"👹 ogre",
	"👽 paul",
	"🦄 pegasus",
	"💩 poo",
	"🤖 rusty",
	"💀 skeletor",
	"🐉 smaug",
	"🧛‍♀️ vampire",
	"👾 xenomorph",
}

type TestJobInput struct {
	sequenceNo int // allocated by observer
	Recipient  string
}

func (i TestJobInput) SequenceNo() int {
	return i.sequenceNo
}

type TestJobResult = string
type TestResultStream chan async.JobResult[TestJobResult]

var greeter = func(j async.Job[TestJobInput]) (async.JobResult[TestJobResult], error) {
	r := rand.Intn(1000) + 1 //nolint:gosec // trivial
	delay := time.Millisecond * time.Duration(r)
	time.Sleep(delay)

	result := async.JobResult[TestJobResult]{
		Payload: fmt.Sprintf("			---> 🍉🍉🍉 [Seq: %v] Hello: '%v'",
			j.Input.SequenceNo(), j.Input.Recipient,
		),
	}

	return result, nil
}

type pipeline[I, R any] struct {
	wg        sync.WaitGroup
	sequence  int
	resultsCh chan async.JobResult[R]
	provider  helpers.ProviderFunc[I]
	producer  *helpers.Producer[I, R]
	pool      *async.WorkerPool[I, R]
	consumer  *helpers.Consumer[R]
	cancel    TerminatorFunc[I, R]
	stop      TerminatorFunc[I, R]
}

func start[I, R any]() *pipeline[I, R] {
	pipe := &pipeline[I, R]{
		resultsCh: make(chan async.JobResult[R], ResultChSize),
		stop: func(_ context.Context, _ time.Duration, _ ...context.CancelFunc) {
			// no-op
		},
		cancel: func(_ context.Context, _ time.Duration, _ ...context.CancelFunc) {
			// no-op
		},
	}

	return pipe
}

func (p *pipeline[I, R]) produce(ctx context.Context, provider helpers.ProviderFunc[I]) {
	p.cancel = func(ctx context.Context, delay time.Duration, cancellations ...context.CancelFunc) {
		go helpers.CancelProducerAfter[I, R](
			delay,
			cancellations...,
		)
	}
	p.stop = func(ctx context.Context, delay time.Duration, _ ...context.CancelFunc) {
		go helpers.StopProducerAfter(
			ctx,
			p.producer,
			delay,
		)
	}

	p.producer = helpers.StartProducer[I, R](
		ctx,
		&p.wg,
		JobChSize,
		provider,
		Delay,
	)

	p.wg.Add(1)
}

func (p *pipeline[I, R]) process(ctx context.Context, noWorkers int, executive async.ExecutiveFunc[I, R]) {
	p.pool = async.NewWorkerPool[I, R](
		&async.NewWorkerPoolParams[I, R]{
			NoWorkers: noWorkers,
			Exec:      executive,
			JobsCh:    p.producer.JobsCh,
			CancelCh:  make(async.CancelStream),
			Quit:      &p.wg,
		})

	go p.pool.Start(ctx, p.resultsCh)

	p.wg.Add(1)
}

func (p *pipeline[I, R]) consume(ctx context.Context) {
	p.consumer = helpers.StartConsumer(ctx,
		&p.wg,
		p.resultsCh,
	)

	p.wg.Add(1)
}

var _ = Describe("WorkerPool", func() {
	When("given: a stream of jobs", func() {
		Context("and: Stopped", func() {
			It("🧪 should: receive and process all", func(ctx SpecContext) {
				defer leaktest.Check(GinkgoT())()
				pipe := start[TestJobInput, TestJobResult]()

				By("👾 WAIT-GROUP ADD(producer)")
				sequence := 0
				pipe.produce(ctx, func() TestJobInput {
					recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
					sequence++
					return TestJobInput{
						sequenceNo: sequence,
						Recipient:  audience[recipient],
					}
				})

				By("👾 WAIT-GROUP ADD(worker-pool)\n")
				pipe.process(ctx, DefaultNoWorkers, greeter)

				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(ctx)

				By("👾 NOW AWAITING TERMINATION")
				pipe.stop.After(ctx, time.Second/5)
				pipe.wg.Wait()

				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					pipe.producer.Count,
					pipe.consumer.Count,
				)

				Expect(pipe.producer.Count).To(Equal(pipe.consumer.Count))
				Eventually(ctx, pipe.resultsCh).WithTimeout(time.Second * 5).Should(BeClosed())
				Eventually(ctx, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())
			}, SpecTimeout(time.Second*5))
		})

		Context("and: Cancelled", func() {
			It("🧪 should: handle cancellation and shutdown cleanly", func(ctxSpec SpecContext) {
				defer leaktest.Check(GinkgoT())()
				pipe := start[TestJobInput, TestJobResult]()

				ctxCancel, cancel := context.WithCancel(ctxSpec)
				cancellations := []context.CancelFunc{cancel}

				By("👾 WAIT-GROUP ADD(producer)")
				sequence := 0
				pipe.produce(ctxCancel, func() TestJobInput {
					recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
					sequence++
					return TestJobInput{
						sequenceNo: sequence,
						Recipient:  audience[recipient],
					}
				})

				By("👾 WAIT-GROUP ADD(worker-pool)\n")
				pipe.process(ctxCancel, DefaultNoWorkers, greeter)

				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(ctxCancel)

				By("👾 NOW AWAITING TERMINATION")
				pipe.cancel.After(ctxCancel, time.Second/5, cancellations...)

				pipe.wg.Wait()

				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					pipe.producer.Count,
					pipe.consumer.Count,
				)

				// The producer count is higher than the consumer count. As a feature, we could
				// collate the numbers produced vs the numbers consumed and perhaps also calculate
				// which jobs were not processed, each indicated with their corresponding Input
				// value.

				// Eventually(ctxCancel, pipe.resultsCh).WithTimeout(time.Second * 5).Should(BeClosed())
				// Eventually(ctxCancel, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())

			}, SpecTimeout(time.Second*5))
		})
	})
})
