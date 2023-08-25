package async_test

import (
	"context"
	"fmt"
	"math/rand"
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
type TerminatorFunc[I, O any] func(ctx context.Context, delay time.Duration, funcs ...context.CancelFunc)

func (f TerminatorFunc[I, O]) After(ctx context.Context, delay time.Duration, funcs ...context.CancelFunc) {
	f(ctx, delay, funcs...)
}

const (
	JobChSize     = 10
	OutputsChSize = 10
	Delay         = 750
)

var (
	audience = []string{
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

	noOp = func(_ context.Context, _ time.Duration, _ ...context.CancelFunc) {}

	testMain = async.GoRoutineName("👾 test-main")
)

type TestJobInput struct {
	Recipient string
}

type TestJobOutput = string
type TestOutputStream chan async.JobOutput[TestJobOutput]

var greeter = func(j async.Job[TestJobInput]) (async.JobOutput[TestJobOutput], error) {
	r := rand.Intn(1000) + 1 //nolint:gosec // trivial
	delay := time.Millisecond * time.Duration(r)
	time.Sleep(delay)

	result := async.JobOutput[TestJobOutput]{
		Payload: fmt.Sprintf("			---> 🍉🍉🍉 [Seq: %v] Hello: '%v'",
			j.SequenceNo, j.Input.Recipient,
		),
	}

	return result, nil
}

type pipeline[I, O any] struct {
	wgex      async.WaitGroupEx
	sequence  int
	outputsCh chan async.JobOutput[O]
	provider  helpers.ProviderFunc[I]
	producer  *helpers.Producer[I, O]
	pool      *async.WorkerPool[I, O]
	consumer  *helpers.Consumer[O]
	cancel    TerminatorFunc[I, O]
	stop      TerminatorFunc[I, O]
}

func start[I, O any]() *pipeline[I, O] {
	pipe := &pipeline[I, O]{
		wgex:      async.NewAnnotatedWaitGroup("🍂 pipeline"),
		outputsCh: make(chan async.JobOutput[O], OutputsChSize),
		stop:      noOp,
		cancel:    noOp,
	}

	return pipe
}

func (p *pipeline[I, O]) produce(ctx context.Context, provider helpers.ProviderFunc[I]) {
	p.cancel = func(ctx context.Context, delay time.Duration, cancellations ...context.CancelFunc) {
		go helpers.CancelProducerAfter[I, O](
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

	p.producer = helpers.StartProducer[I, O](
		ctx,
		p.wgex,
		JobChSize,
		provider,
		Delay,
	)

	p.wgex.Add(1, p.producer.RoutineName)
}

func (p *pipeline[I, O]) process(ctx context.Context, noWorkers int, executive async.ExecutiveFunc[I, O]) {
	p.pool = async.NewWorkerPool[I, O](
		&async.NewWorkerPoolParams[I, O]{
			NoWorkers: noWorkers,
			Exec:      executive,
			JobsCh:    p.producer.JobsCh,
			CancelCh:  make(async.CancelStream),
			Quitter:   p.wgex,
		})

	go p.pool.Start(ctx, p.outputsCh)

	p.wgex.Add(1, p.pool.RoutineName)
}

func (p *pipeline[I, O]) consume(ctx context.Context) {
	p.consumer = helpers.StartConsumer(ctx,
		p.wgex,
		p.outputsCh,
	)

	p.wgex.Add(1, p.consumer.RoutineName)
}

var _ = Describe("WorkerPool", func() {
	When("given: a stream of jobs", func() {
		Context("and: Stopped", func() {
			It("🧪 should: receive and process all", func(ctx SpecContext) {
				defer leaktest.Check(GinkgoT())()

				pipe := start[TestJobInput, TestJobOutput]()

				defer func() {
					if counter, ok := (pipe.wgex).(async.AssistedCounter); ok {
						fmt.Printf("🎈🎈🎈🎈 remaining count: '%v'\n", counter.Count())
					}
				}()

				By("👾 WAIT-GROUP ADD(producer)")
				provider := func() TestJobInput {
					recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
					return TestJobInput{
						Recipient: audience[recipient],
					}
				}
				pipe.produce(ctx, provider)

				By("👾 WAIT-GROUP ADD(worker-pool)\n")
				pipe.process(ctx, DefaultNoWorkers, greeter)

				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(ctx)

				By("👾 NOW AWAITING TERMINATION")
				pipe.stop.After(ctx, time.Second/5)
				pipe.wgex.Wait(async.GoRoutineName("👾 test-main"))

				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					pipe.producer.Count,
					pipe.consumer.Count,
				)

				Expect(pipe.producer.Count).To(Equal(pipe.consumer.Count))
				Eventually(ctx, pipe.outputsCh).WithTimeout(time.Second * 5).Should(BeClosed())
				Eventually(ctx, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())
			}, SpecTimeout(time.Second*5))
		})

		Context("and: Cancelled", func() {
			It("🧪 should: handle cancellation and shutdown cleanly", func(ctxSpec SpecContext) {
				defer leaktest.Check(GinkgoT())()
				pipe := start[TestJobInput, TestJobOutput]()

				ctxCancel, cancel := context.WithCancel(ctxSpec)
				cancellations := []context.CancelFunc{cancel}

				By("👾 WAIT-GROUP ADD(producer)")
				pipe.produce(ctxCancel, func() TestJobInput {
					recipient := rand.Intn(len(audience)) //nolint:gosec // trivial

					return TestJobInput{
						Recipient: audience[recipient],
					}
				})

				By("👾 WAIT-GROUP ADD(worker-pool)\n")
				pipe.process(ctxCancel, DefaultNoWorkers, greeter)

				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(ctxCancel)

				By("👾 NOW AWAITING TERMINATION")
				pipe.cancel.After(ctxCancel, time.Second/5, cancellations...)

				pipe.wgex.Wait(async.GoRoutineName("👾 test-main"))

				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					pipe.producer.Count,
					pipe.consumer.Count,
				)

				// The producer count is higher than the consumer count. As a feature, we could
				// collate the numbers produced vs the numbers consumed and perhaps also calculate
				// which jobs were not processed, each indicated with their corresponding Input
				// value.

				// Eventually(ctxCancel, pipe.outputsCh).WithTimeout(time.Second * 5).Should(BeClosed())
				// Eventually(ctxCancel, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())

			}, SpecTimeout(time.Second*5))
		})
	})
})
