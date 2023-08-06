package async_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/lorax/async"
	"github.com/snivilised/lorax/internal/helpers"
)

const (
	JobChSize    = 10
	ResultChSize = 10
	Delay        = 750
)

type TestJobInput struct {
	sequenceNo int // allocated by observer
	Recipient  string
}

func (i TestJobInput) SequenceNo() int {
	return i.sequenceNo
}

type TestJobResult = string
type TestResultChan chan async.JobResult[TestJobResult]

type exec struct {
}

func (e *exec) Invoke(j async.Job[TestJobInput]) (async.JobResult[TestJobResult], error) {
	r := rand.Intn(1000) + 1 //nolint:gosec // trivial
	delay := time.Millisecond * time.Duration(r)
	time.Sleep(delay)

	result := async.JobResult[TestJobResult]{
		Payload: fmt.Sprintf("	---> exec.Invoke [Seq: %v]🍉 Hello: '%v'",
			j.Input.SequenceNo(), j.Input.Recipient,
		),
	}
	fmt.Println(result.Payload)

	return result, nil
}

var _ = Describe("WorkerPool", func() {
	Context("producer/consumer", func() {
		When("given: a stream of jobs", func() {
			It("🧪 should: receive and process all", func(specCtx SpecContext) {
				var (
					wg sync.WaitGroup
				)
				sequence := 0

				resultsCh := make(chan async.JobResult[TestJobResult], ResultChSize)

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(producer)")

				provider := func() TestJobInput {
					sequence++
					return TestJobInput{
						sequenceNo: sequence,
						Recipient:  "jimmy 🦊",
					}
				}

				producer := helpers.NewProducer[TestJobInput, TestJobResult](specCtx, &wg, JobChSize, provider, Delay)
				pool := async.NewWorkerPool[TestJobInput, TestJobResult](&async.NewWorkerPoolParams[TestJobInput, TestJobResult]{
					Exec:   &exec{},
					JobsCh: producer.JobsCh,
					Cancel: make(async.CancelStream),
					Quit:   &wg,
				})

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(worker-pool)\n")

				go pool.Run(specCtx, resultsCh)

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(consumer)")

				consumer := helpers.NewConsumer(specCtx, &wg, resultsCh)

				go func() {
					snooze := time.Second / 5
					fmt.Printf("		>>> 💤 Sleeping before requesting stop (%v) ...\n", snooze)
					time.Sleep(snooze)
					producer.Stop()
					fmt.Printf("		>>> 🍧🍧🍧 stop submitted.\n")
				}()

				wg.Wait()
				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					producer.Count,
					consumer.Count,
				)

				Expect(producer.Count).To(Equal(consumer.Count))
				Eventually(specCtx, resultsCh).WithTimeout(time.Second * 2).Should(BeClosed())
				Eventually(specCtx, producer.JobsCh).WithTimeout(time.Second * 2).Should(BeClosed())
			}, SpecTimeout(time.Second*2))
		})

		When("given: cancellation invoked before end of work", func() {
			XIt("🧪 should: close down gracefully", func(specCtx SpecContext) {
				// this case shows that worker pool needs a redesign. Each worker
				// go routine needs to have a lifetime that spans the lifetime of
				// the session, rather than a short lifetime that matches that of
				// an individual job. This will make processing more reliable,
				// especially when it comes to cancellation. As it is, since the
				// worker GR only exists for the lifetime of the job, when the
				// job is short (in duration), it is very unlikely it will see
				// the cancellation request and therefore and therefore likely
				// to send to a closed channel (the result channel).
				//
				var (
					wg sync.WaitGroup
				)
				sequence := 0

				resultsCh := make(chan async.JobResult[TestJobResult], ResultChSize)

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(producer)")

				provider := func() TestJobInput {
					sequence++
					return TestJobInput{
						sequenceNo: sequence,
						Recipient:  "johnny 😈",
					}
				}
				ctx, cancel := context.WithCancel(specCtx)

				producer := helpers.NewProducer[TestJobInput, TestJobResult](ctx, &wg, JobChSize, provider, Delay)
				pool := async.NewWorkerPool[TestJobInput, TestJobResult](&async.NewWorkerPoolParams[TestJobInput, TestJobResult]{
					Exec:   &exec{},
					JobsCh: producer.JobsCh,
					Cancel: make(async.CancelStream),
					Quit:   &wg,
				})

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(worker-pool)\n")

				go pool.Run(ctx, resultsCh)

				wg.Add(1)
				By("👾 WAIT-GROUP ADD(consumer)")

				consumer := helpers.NewConsumer(ctx, &wg, resultsCh)

				go func() {
					snooze := time.Second / 10
					fmt.Printf("		>>> 💤 Sleeping before requesting cancellation (%v) ...\n", snooze)
					time.Sleep(snooze)
					cancel()
					fmt.Printf("		>>> 🍧🍧🍧 cancel submitted.\n")
				}()

				wg.Wait()
				fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
					producer.Count,
					consumer.Count,
				)

				Eventually(specCtx, resultsCh).WithTimeout(time.Second * 2).Should(BeClosed())
				Eventually(specCtx, producer.JobsCh).WithTimeout(time.Second * 2).Should(BeClosed())
			}, SpecTimeout(time.Second*2))
		})
	})

	Context("ginkgo consumer", func() {
		It("🧪 should: receive and process all", func(specCtx SpecContext) {
			var (
				wg sync.WaitGroup
			)
			sequence := 0

			resultsCh := make(chan async.JobResult[TestJobResult], ResultChSize)

			wg.Add(1)
			By("👾 WAIT-GROUP ADD(producer)")

			provider := func() TestJobInput {
				sequence++
				return TestJobInput{
					sequenceNo: sequence,
					Recipient:  "cosmo 👽",
				}
			}

			producer := helpers.NewProducer[TestJobInput, TestJobResult](specCtx, &wg, JobChSize, provider, Delay)
			pool := async.NewWorkerPool[TestJobInput, TestJobResult](&async.NewWorkerPoolParams[TestJobInput, TestJobResult]{
				Exec:   &exec{},
				JobsCh: producer.JobsCh,
				Cancel: make(async.CancelStream),
				Quit:   &wg,
			})

			wg.Add(1)
			By("👾 WAIT-GROUP ADD(worker-pool)\n")

			go pool.Run(specCtx, resultsCh)

			wg.Add(1)
			By("👾 WAIT-GROUP ADD(consumer)")

			consumer := helpers.NewConsumer(specCtx, &wg, resultsCh)

			go func() {
				snooze := time.Second / 5
				fmt.Printf("		>>> 💤 Sleeping before requesting stop (%v) ...\n", snooze)
				time.Sleep(snooze)
				producer.Stop()
				fmt.Printf("		>>> 🍧🍧🍧 stop submitted.\n")
			}()

			wg.Wait()
			fmt.Printf("<--- orpheus(alpha) finished Counts >>> (Producer: '%v', Consumer: '%v'). 🎯🎯🎯\n",
				producer.Count,
				consumer.Count,
			)

			Expect(producer.Count).To(Equal(consumer.Count))
			Eventually(specCtx, resultsCh).WithTimeout(time.Second * 2).Should(BeClosed())
			Eventually(specCtx, producer.JobsCh).WithTimeout(time.Second * 2).Should(BeClosed())
		}, SpecTimeout(time.Second*2))
	})
})
