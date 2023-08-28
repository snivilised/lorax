package async_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"

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

type TestJobOutput string
type TestOutputStream chan async.JobOutput[TestJobOutput]

var greeter = func(j async.Job[TestJobInput]) (async.JobOutput[TestJobOutput], error) {
	r := rand.Intn(1000) + 1 //nolint:gosec // trivial
	delay := time.Millisecond * time.Duration(r)
	time.Sleep(delay)

	result := async.JobOutput[TestJobOutput]{
		Payload: TestJobOutput(fmt.Sprintf("			---> 🍉🍉🍉 [Seq: %v] Hello: '%v'",
			j.SequenceNo, j.Input.Recipient,
		)),
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

func start[I, O any](outputsCh chan async.JobOutput[O]) *pipeline[I, O] {
	pipe := &pipeline[I, O]{
		wgex:      async.NewAnnotatedWaitGroup("🍂 pipeline"),
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

type TestPipeline *pipeline[TestJobInput, TestJobOutput]
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
				func() chan async.JobOutput[TestJobOutput] {
					return make(chan async.JobOutput[TestJobOutput], entry.outputsChSize)
				},
				func() chan async.JobOutput[TestJobOutput] {
					return nil
				},
			)
			pipe := start[TestJobInput, TestJobOutput](oc)

			defer func() {
				if counter, ok := (pipe.wgex).(async.AnnotatedWgCounter); ok {
					fmt.Printf("🎈🎈🎈🎈 remaining count: '%v'\n", counter.Count())
				}
			}()

			ctx, cancel := entry.context(ctxSpec)

			By("👾 WAIT-GROUP ADD(producer)")
			provider := func() TestJobInput {
				recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
				return TestJobInput{
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
			pipe.wgex.Wait(async.GoRoutineName("👾 test-main"))

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
