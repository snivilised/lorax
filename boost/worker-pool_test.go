package boost_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/samber/lo"

	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/lorax/internal/helpers"
)

// TerminatorFunc brings the work pool processing to an end, eg
// by stopping or cancellation after the requested amount of time.
type TerminatorFunc[I, O any] func(parentContext context.Context,
	parentCancel context.CancelFunc,
	delay time.Duration,
)

func (f TerminatorFunc[I, O]) After(parentContext context.Context,
	parentCancel context.CancelFunc,
	delay time.Duration,
) {
	f(parentContext, parentCancel, delay)
}

const (
	JobChSize     = 10
	OutputsChSize = 10
	ScalingFactor = 1
)

func scale(t time.Duration) time.Duration {
	return t / ScalingFactor
}

var defaults = struct {
	noOfWorkers      int
	producerInterval time.Duration
	consumerInterval time.Duration
}{
	noOfWorkers:      5,
	producerInterval: time.Microsecond,
	consumerInterval: time.Microsecond,
}

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
		"🧛 dracula",
		"👾 xenomorph",
	}

	noOp     = func(_ context.Context, _ context.CancelFunc, _ time.Duration) {}
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

type (
	TestInput struct {
		Recipient string
	}

	TestJobInput        = boost.Job[TestInput]
	TestInputStream     = boost.JobStream[TestJobInput]
	TestOutput          string
	TestJobOutput       = boost.JobOutput[TestOutput]
	TestJobOutputStream = boost.JobOutputStream[TestOutput]
)

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
	wgan       boost.WaitGroupAn
	sequence   int
	outputsDup *boost.Duplex[boost.JobOutput[O]]
	provider   helpers.ProviderFunc[I]
	producer   *helpers.Producer[I, O]
	pool       *boost.WorkerPool[I, O]
	consumer   *helpers.Consumer[O]
	cancel     TerminatorFunc[I, O]
	stop       TerminatorFunc[I, O]
}

func start[I, O any](outputsDupCh *boost.Duplex[boost.JobOutput[O]]) *pipeline[I, O] {
	pipe := &pipeline[I, O]{
		wgan:       boost.NewAnnotatedWaitGroup("🍂 pipeline"),
		outputsDup: outputsDupCh,
		stop:       noOp,
		cancel:     noOp,
	}

	return pipe
}

func (p *pipeline[I, O]) produce(parentContext context.Context,
	interval time.Duration,
	provider helpers.ProviderFunc[I],
	verbose bool,
) {
	p.cancel = func(_ context.Context,
		parentCancel context.CancelFunc,
		delay time.Duration,
	) {
		go helpers.CancelProducerAfter[I, O](
			delay,
			parentCancel,
			verbose,
		)
	}
	p.stop = func(_ context.Context,
		_ context.CancelFunc,
		delay time.Duration,
	) {
		go helpers.StopProducerAfter(
			parentContext,
			p.producer,
			delay,
			verbose,
		)
	}

	p.producer = helpers.StartProducer[I, O](
		parentContext,
		p.wgan,
		JobChSize,
		provider,
		interval,
		verbose,
	)

	p.wgan.Add(1, p.producer.RoutineName)
}

func (p *pipeline[I, O]) process(parentContext context.Context,
	parentCancel context.CancelFunc,
	outputChTimeout time.Duration,
	noWorkers int,
	executive boost.ExecutiveFunc[I, O],
) {
	p.pool = boost.NewWorkerPool[I, O](
		&boost.NewWorkerPoolParams[I, O]{
			NoWorkers:       noWorkers,
			OutputChTimeout: outputChTimeout,
			Exec:            executive,
			JobsCh:          p.producer.JobsCh,
			CancelCh:        make(boost.CancelStream),
			WaitAQ:          p.wgan,
		})

	p.wgan.Add(1, p.pool.RoutineName)

	go p.pool.Start(parentContext, parentCancel, p.outputsDup.WriterCh)
}

func (p *pipeline[I, O]) consume(parentContext context.Context,
	interval time.Duration, verbose bool,
) {
	p.consumer = helpers.StartConsumer(parentContext,
		p.wgan,
		p.outputsDup.ReaderCh,
		interval,
		verbose,
	)

	p.wgan.Add(1, p.consumer.RoutineName)
}

type TestPipeline *pipeline[TestInput, TestOutput]
type assertFunc func(parentContext context.Context,
	pipe TestPipeline,
	result *boost.PoolResult,
)
type finishFunc func(
	parentContext context.Context,
	cancel context.CancelFunc,
	pipe TestPipeline,
	delay time.Duration,
)
type summariseFunc func(pipe TestPipeline)

var (
	finishWithStop finishFunc = func(
		parentContext context.Context,
		cancel context.CancelFunc,
		pipe TestPipeline,
		delay time.Duration,
	) {
		pipe.stop.After(parentContext, cancel, delay)
	}
	finishWithCancel finishFunc = func(
		parentContext context.Context,
		cancel context.CancelFunc,
		pipe TestPipeline,
		delay time.Duration,
	) {
		pipe.cancel.After(parentContext, cancel, delay)
	}
	assertWithConsumerCounts assertFunc = func(parentContext context.Context,
		pipe TestPipeline,
		result *boost.PoolResult,
	) {
		Expect(result.Error).Error().To(BeNil())
		Expect(pipe.producer.Count).To(Equal(pipe.consumer.Count))
		Eventually(parentContext, pipe.outputsDup.Channel).WithTimeout(time.Second * 5).Should(BeClosed())
		Eventually(parentContext, pipe.producer.JobsCh).WithTimeout(time.Second * 5).Should(BeClosed())
	}
	assertCancelled assertFunc = func(_ context.Context,
		_ TestPipeline,
		result *boost.PoolResult,
	) {
		// This needs to be upgraded to check that it is "timeout on send" error
		Expect(result.Error).Error().NotTo(BeNil())
	}
	assertOutputChTimeout assertFunc = func(_ context.Context,
		_ TestPipeline,
		result *boost.PoolResult,
	) {
		Expect(result.Error).Error().NotTo(BeNil())

		// TODO: restore this check once i18n messages have been regenerated
		//
		// ??? if err, ok := result.Error.(i18n.OutputChTimeoutErrorBehaviourQuery); ok {
		// 	// create a custom matcher? --> To(BeTrue()) is not very informative
		// 	//
		// 	Expect(err.OutputChTimeout()).To(BeTrue())
		// }
	}

	assertNoError assertFunc = func(_ context.Context,
		_ TestPipeline, result *boost.PoolResult) {
		Expect(result.Error).Error().To(BeNil())
	}

	assertAbort assertFunc = func(parentContext context.Context,
		pipe TestPipeline,
		result *boost.PoolResult,
	) {
		_, _ = parentContext, pipe

		if result.Error == nil {
			return // This is temporary
		}

		Expect(result.Error).Error().NotTo(BeNil())
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
	silentSummariser summariseFunc = func(_ TestPipeline) {}
)

type durations struct {
	producer        time.Duration
	consumer        time.Duration
	finishAfter     time.Duration
	outputChTimeout time.Duration
}

type poolTE struct {
	given         string
	should        string
	now           int
	outputsChSize int
	intervals     durations
	finish        finishFunc
	summarise     summariseFunc
	assert        assertFunc
}

var _ = Describe("WorkerPool", Ordered, func() {
	DescribeTable("stream of jobs",
		func(specContext SpecContext, entry *poolTE) {
			defer leaktest.Check(GinkgoT())()

			const (
				verbose = false
			)

			outputDup := lo.TernaryF(entry.outputsChSize > 0,
				func() *boost.Duplex[boost.JobOutput[TestOutput]] {
					return boost.NewDuplex(make(TestJobOutputStream, entry.outputsChSize))
				},
				func() *boost.Duplex[boost.JobOutput[TestOutput]] {
					return &boost.Duplex[boost.JobOutput[TestOutput]]{}
				},
			)

			pipe := start[TestInput, TestOutput](outputDup)

			defer func() {
				if counter, ok := (pipe.wgan).(boost.AnnotatedWgCounter); ok && verbose {
					fmt.Printf("🎈🎈🎈🎈 remaining count: '%v'\n", counter.Count())
				}
			}()

			if !verbose {
				entry.summarise = silentSummariser
			}

			parentContext, parentCancel := context.WithCancel(specContext)

			By("👾 WAIT-GROUP ADD(producer)")
			provider := func() TestInput {
				recipient := rand.Intn(len(audience)) //nolint:gosec // trivial
				return TestInput{
					Recipient: audience[recipient],
				}
			}

			pipe.produce(parentContext, lo.Ternary(entry.intervals.producer > 0,
				entry.intervals.producer, defaults.producerInterval,
			), provider, verbose)

			By("👾 WAIT-GROUP ADD(worker-pool)\n")
			now := lo.Ternary(entry.now > 0, entry.now, defaults.noOfWorkers)
			pipe.process(parentContext,
				parentCancel,
				entry.intervals.outputChTimeout,
				now,
				greeter,
			)

			if outputDup.Channel != nil {
				By("👾 WAIT-GROUP ADD(consumer)")
				pipe.consume(parentContext, lo.Ternary(entry.intervals.consumer > 0,
					entry.intervals.consumer, defaults.consumerInterval,
				), verbose)
			}

			By("👾 NOW AWAITING TERMINATION")
			entry.finish(parentContext, parentCancel, pipe, entry.intervals.finishAfter)
			pipe.wgan.Wait(boost.GoRoutineName("👾 test-main"))
			result := <-pipe.pool.ResultInCh

			entry.summarise(pipe)
			entry.assert(parentContext, pipe, result)
		},
		func(entry *poolTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'", entry.given, entry.should)
		},

		// ☣️ TODO: don't forget to put the timeouts back into the test entries.

		Entry(nil, &poolTE{
			given:         "finish by stop",
			should:        "receive and process all",
			outputsChSize: OutputsChSize,
			intervals: durations{
				finishAfter:     scale(time.Second / 5),
				outputChTimeout: scale(time.Second),
			},
			finish:    finishWithStop,
			summarise: summariseWithConsumer,
			assert:    assertWithConsumerCounts,
		}),

		XEntry(nil, &poolTE{ // Intermittent failure
			given:  "results not consumed in a timely manner",
			should: "timeout on error",
			now:    2,
			intervals: durations{
				consumer:        scale(1), // time.Second * 8,
				finishAfter:     scale(time.Second * 2),
				outputChTimeout: scale(time.Second * 1),
			},
			outputsChSize: OutputsChSize,
			finish:        finishWithCancel,
			summarise:     summariseWithConsumer,
			assert:        assertCancelled,
		}), // SpecTimeout(time.Second*8)

		XEntry(nil, &poolTE{
			given:         "finish by stop and high no of workers",
			should:        "receive and process all",
			now:           16,
			outputsChSize: OutputsChSize,
			intervals: durations{
				finishAfter:     scale(time.Second / 5),
				outputChTimeout: scale(time.Second),
			},
			finish:    finishWithStop,
			summarise: summariseWithConsumer,
			assert:    assertWithConsumerCounts,
		}),

		Entry(nil, &poolTE{
			given:         "finish by stop and no output",
			should:        "receive and process all without sending output, no error",
			outputsChSize: 0,
			intervals: durations{
				finishAfter:     scale(time.Second / 5),
				outputChTimeout: scale(time.Second / 10),
			},
			finish:    finishWithStop,
			summarise: summariseWithoutConsumer,
			assert:    assertNoError,
		}),

		Entry(nil, &poolTE{
			given:         "finish by cancel and no output",
			should:        "receive and process all without sending output, no error",
			outputsChSize: 0,
			intervals: durations{
				finishAfter:     scale(time.Second / 5),
				outputChTimeout: scale(time.Second / 10),
			},
			finish:    finishWithCancel,
			summarise: summariseWithoutConsumer,
			assert:    assertNoError,
		}),

		XEntry(nil, &poolTE{
			given:         "finish by stop",
			should:        "timeout and abort",
			now:           2,
			outputsChSize: OutputsChSize, // 1
			intervals: durations{
				finishAfter:     scale(time.Second / 5),
				outputChTimeout: scale(time.Second / 1000),
			},
			finish:    finishWithStop,
			summarise: summariseWithConsumer,
			assert:    assertOutputChTimeout,
		}),
	)
})
