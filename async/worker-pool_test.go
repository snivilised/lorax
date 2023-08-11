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

func init() { rand.Seed(time.Now().Unix()) }

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
	provider  helpers.ProviderFn[I]
	producer  *helpers.Producer[I, R]
	pool      *async.WorkerPool[I, R]
	consumer  *helpers.Consumer[R]
}

func start[I, R any]() *pipeline[I, R] {
	resultsCh := make(chan async.JobResult[R], ResultChSize)

	pipe := &pipeline[I, R]{
		resultsCh: resultsCh,
	}

	return pipe
}

func (p *pipeline[I, R]) startProducer(ctx context.Context, provider helpers.ProviderFn[I]) {
	p.producer = helpers.StartProducer[I, R](
		ctx,
		&p.wg,
		JobChSize,
		provider,
		Delay,
	)

	p.wg.Add(1)
}

func (p *pipeline[I, R]) startPool(ctx context.Context, executive async.ExecutiveFunc[I, R]) {
	p.pool = async.NewWorkerPool[I, R](
		&async.NewWorkerPoolParams[I, R]{
			NoWorkers: 5,
			Exec:      executive,
			JobsCh:    p.producer.JobsCh,
			CancelCh:  make(async.CancelStream),
			Quit:      &p.wg,
		})

	go p.pool.Start(ctx, p.resultsCh)

	p.wg.Add(1)
}

func (p *pipeline[I, R]) startConsumer(ctx context.Context) {
	p.consumer = helpers.StartConsumer(ctx,
		&p.wg,
		p.resultsCh,
	)

	p.wg.Add(1)
}

func (p *pipeline[I, R]) stopProducerAfter(ctx context.Context, after time.Duration) {
	go helpers.StopProducerAfter(
		ctx,
		p.producer,
		after,
	)
}

var _ = Describe("WorkerPool", func() {
	When("given: a stream of jobs", func() {
		Context("and: Stopped", func() {
			It("🧪 should: receive and process all", func(ctx SpecContext) {
				defer leaktest.Check(GinkgoT())()
				pipe := start[TestJobInput, TestJobResult]()

				By("👾 WAIT-GROUP ADD(producer)")
				sequence := 0
				pipe.startProducer(ctx, func() TestJobInput {
					recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
					sequence++
					return TestJobInput{
						sequenceNo: sequence,
						Recipient:  audience[recipient],
					}
				})

				By("👾 WAIT-GROUP ADD(worker-pool)\n")
				pipe.startPool(ctx, greeter)

				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.startConsumer(ctx)

				By("👾 NOW AWAITING TERMINATION")
				pipe.stopProducerAfter(ctx, time.Second/5)
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
			It("should test something", func() { // It is ginkgo test case function
				Expect(audience).To(HaveLen(14))
			})

			It("🧪 should: handle cancellation and shutdown cleanly", func(_ SpecContext) {

			})
		})
	})
})
