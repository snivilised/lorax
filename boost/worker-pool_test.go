package boost_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

	"github.com/snivilised/lorax/boost"
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

	testMain = boost.GoRoutineName("👾 test-main")
)

// When defining client side channel types, the rule should be, when creating
// a derivative of the boost type, the client should not introduce a new type,
// rather, they should introduce an alias to the boost type. So we should never
// do:
// type TestInputStream chan boost.Job[TestJobInput]
//
// because we are referring to a boost type. Instead we should define
//
// type TestInputStream = chan boost.Job[TestJobInput]
//

type TestInput struct {
	Recipient string
}
type TestJobInput = boost.Job[TestInput]
type TestInputStream = chan boost.Job[TestJobInput]

type TestOutput string
type TestJobOutput = boost.JobOutput[TestOutput]
type TestOutputStream = chan boost.JobOutput[TestOutput]

var greeter = func(j TestJobInput) (TestJobOutput, error) {
	r := rand.Intn(1000) + 1 //nolint:gosec // trivial
	delay := time.Millisecond * time.Duration(r)
	time.Sleep(delay)

	result := TestJobOutput{
		Payload: TestOutput(fmt.Sprintf("			---> 🍉🍉🍉 [Seq: %v] Hello: '%v'",
			j.SequenceNo, j.Input.Recipient,
		)),
	}

	return result, nil
}

type pipeline[I, O any] struct {
	wgan      boost.WaitGroupAn
	sequence  int
	outputsCh chan boost.JobOutput[O]
	provider  helpers.ProviderFunc[I]
	producer  *helpers.Producer[I, O]
	pool      *boost.WorkerPool[I, O]
	consumer  *helpers.Consumer[O]
	cancel    TerminatorFunc[I, O]
	stop      TerminatorFunc[I, O]
}

func start[I, O any](outputsCh chan boost.JobOutput[O]) *pipeline[I, O] {
	pipe := &pipeline[I, O]{
		wgan:      boost.NewAnnotatedWaitGroup("🍂 pipeline"),
		outputsCh: outputsCh,
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
		p.wgan,
		JobChSize,
		provider,
		Delay,
	)

	p.wgan.Add(1, p.producer.RoutineName)
}

func (p *pipeline[I, O]) process(ctx context.Context, noWorkers int, executive boost.ExecutiveFunc[I, O]) {
	p.pool = boost.NewWorkerPool[I, O](
		&boost.NewWorkerPoolParams[I, O]{
			NoWorkers: noWorkers,
			Exec:      executive,
			JobsCh:    p.producer.JobsCh,
			CancelCh:  make(boost.CancelStream),
			WaitAQ:    p.wgan,
		})

	go p.pool.Start(ctx, p.outputsCh)

	p.wgan.Add(1, p.pool.RoutineName)
}

func (p *pipeline[I, O]) consume(ctx context.Context) {
	p.consumer = helpers.StartConsumer(ctx,
		p.wgan,
		p.outputsCh,
	)

	p.wgan.Add(1, p.consumer.RoutineName)
}

type TestPipeline *pipeline[TestInput, TestOutput]
type assertFunc func(ctx context.Context, pipe TestPipeline)
type contextFunc func(ctx context.Context) (context.Context, context.CancelFunc)
type finishFunc func(
	ctx context.Context,
	pipe TestPipeline,
	delay time.Duration,
	cancel context.CancelFunc,
)
type summariseFunc func(pipe TestPipeline)

var (
	finishWithStop finishFunc = func(
		ctx context.Context,
		pipe TestPipeline,
		delay time.Duration,
		cancel context.CancelFunc,
	) {
		pipe.stop.After(ctx, delay)
	}
	finishWithCancel finishFunc = func(
		ctx context.Context,
		pipe TestPipeline,
		delay time.Duration,
		cancel context.CancelFunc,
	) {
		pipe.cancel.After(ctx, delay, cancel)
	}
	passthruContext contextFunc = func(ctx context.Context) (context.Context, context.CancelFunc) {
		return ctx, nil
	}
	assertCounts assertFunc = func(ctx context.Context, pipe TestPipeline) {
		Expect(pipe.producer.Count).To(Equal(pipe.consumer.Count))
		Eventually(ctx, pipe.outputsCh).WithTimeout(time.Second * 5).Should(BeClosed())
		Eventually(ctx, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())
	}
	summariseWithConsumer summariseFunc = func(pipe TestPipeline) {
		fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
			pipe.producer.Count,
			pipe.consumer.Count,
		)
	}
	summariseWithoutConsumer summariseFunc = func(pipe TestPipeline) {
		fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', NO CONSUMER). 🎯🎯🎯\n",
			pipe.producer.Count,
		)
	}
)

type poolTE struct {
	given         string
	should        string
	now           int
	outputsChSize int
	after         time.Duration
	context       contextFunc
	finish        finishFunc
	summarise     summariseFunc
	assert        assertFunc
}

var _ = Describe("WorkerPool", func() {
	DescribeTable("stream of jobs",
		func(ctxSpec SpecContext, entry *poolTE) {
			defer leaktest.Check(GinkgoT())()

			oc := lo.TernaryF(entry.outputsChSize > 0,
				func() TestOutputStream {
					return make(TestOutputStream, entry.outputsChSize)
				},
				func() TestOutputStream {
					return nil
				},
			)
			pipe := start[TestInput, TestOutput](oc)

			defer func() {
				if counter, ok := (pipe.wgan).(boost.AnnotatedWgCounter); ok {
					fmt.Printf("🎈🎈🎈🎈 remaining count: '%v'\n", counter.Count())
				}
			}()

			ctx, cancel := entry.context(ctxSpec)

			By("👾 WAIT-GROUP ADD(producer)")
			provider := func() TestInput {
				recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
				return TestInput{
					Recipient: audience[recipient],
				}
			}
			pipe.produce(ctx, provider)

			By("👾 WAIT-GROUP ADD(worker-pool)\n")
			now := lo.Ternary(entry.now > 0, entry.now, DefaultNoWorkers)
			pipe.process(ctx, now, greeter)

			if oc != nil {
				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(ctx)
			}

			By("👾 NOW AWAITING TERMINATION")
			entry.finish(ctx, pipe, entry.after, cancel)
			pipe.wgan.Wait(boost.GoRoutineName("👾 test-main"))

			entry.summarise(pipe)
			if entry.assert != nil {
				entry.assert(ctx, pipe)
			}
		},
		func(entry *poolTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'", entry.given, entry.should)
		},

		Entry(nil, &poolTE{
			given:         "finish by stop",
			should:        "receive and process all",
			outputsChSize: OutputsChSize,
			after:         time.Second / 5,
			context:       passthruContext,
			finish:        finishWithStop,
			summarise:     summariseWithConsumer,
			assert:        assertCounts,
		}),

		Entry(nil, &poolTE{
			given:         "finish by cancel",
			should:        "receive and process all",
			outputsChSize: OutputsChSize,
			after:         time.Second / 5,
			context:       context.WithCancel,
			finish:        finishWithCancel,
			summarise:     summariseWithConsumer,
		}),

		Entry(nil, &poolTE{
			given:         "finish by stop and no output",
			should:        "receive and process all",
			outputsChSize: 0,
			after:         time.Second / 5,
			context:       passthruContext,
			finish:        finishWithStop,
			summarise:     summariseWithoutConsumer,
		}),

		Entry(nil, &poolTE{
			given:         "finish by cancel and no output",
			should:        "receive and process all",
			outputsChSize: 0,
			after:         time.Second / 5,
			context:       context.WithCancel,
			finish:        finishWithCancel,
			summarise:     summariseWithoutConsumer,
		}),

		Entry(nil, &poolTE{
			given:         "finish by stop and high no of workers",
			should:        "receive and process all",
			now:           16,
			outputsChSize: OutputsChSize,
			after:         time.Second / 5,
			context:       passthruContext,
			finish:        finishWithStop,
			summarise:     summariseWithConsumer,
			assert:        assertCounts,
		}),
	)
})
