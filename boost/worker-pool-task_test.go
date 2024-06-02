package boost_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/lorax/internal/ants"
)

var _ = Describe("WorkerPoolTask", func() {
	Context("ants", func() {
		When("NonBlocking", func() {
			It("should: not fail", func(specCtx SpecContext) {
				// TestNonblockingSubmit
				var wg sync.WaitGroup

				ctx, cancel := context.WithCancel(specCtx)
				defer cancel()

				pool, err := boost.NewTaskPool[int, int](ctx, PoolSize, &wg,
					boost.WithNonblocking(true),
				)
				defer pool.Release(ctx)

				Expect(err).To(Succeed())
				Expect(pool).NotTo(BeNil())

				for i := 0; i < PoolSize-1; i++ {
					Expect(pool.Post(ctx, longRunningFunc)).To(Succeed(),
						"nonblocking submit when pool is not full shouldn't return error",
					)
				}

				firstCh := make(chan struct{})
				secondCh := make(chan struct{})
				fn := func() {
					<-firstCh
					close(secondCh)
				}
				// pool is full now.
				Expect(pool.Post(ctx, fn)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)
				Expect(pool.Post(ctx, demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
					"nonblocking submit when pool is full should get an ErrPoolOverload",
				)
				// interrupt fn to get an available worker
				close(firstCh)
				<-secondCh
				Expect(pool.Post(ctx, demoFunc)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)
			})
		})

		When("MaxNonBlocking", func() {
			It("should: not fail", func(specCtx SpecContext) {
				// TestMaxBlockingSubmit
				var wg sync.WaitGroup

				ctx, cancel := context.WithCancel(specCtx)
				defer cancel()

				pool, err := boost.NewTaskPool[int, int](ctx, PoolSize, &wg,
					boost.WithMaxBlockingTasks(1),
				)
				Expect(err).To(Succeed(), "create TimingPool failed")
				defer pool.Release(ctx)

				By("👾 POOL-CREATED\n")
				for i := 0; i < PoolSize-1; i++ {
					Expect(pool.Post(ctx, longRunningFunc)).To(Succeed(),
						"submit when pool is not full shouldn't return error",
					)
				}
				ch := make(chan struct{})
				fn := func() {
					<-ch
				}
				// pool is full now.
				Expect(pool.Post(ctx, fn)).To(Succeed(),
					"submit when pool is not full shouldn't return error",
				)

				By("👾 WAIT-GROUP ADD(worker-pool-task)\n")
				wg.Add(1)
				errCh := make(chan error, 1)
				go func() {
					// should be blocked. blocking num == 1
					if err := pool.Post(ctx, demoFunc); err != nil {
						errCh <- err
					}
					By("👾 Producer complete\n")
					wg.Done()
				}()

				By("👾 Main sleeping...\n")
				time.Sleep(1 * time.Second)

				// already reached max blocking limit
				Expect(pool.Post(ctx, demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
					"blocking submit when pool reach max blocking submit should return ErrPoolOverload",
				)

				By("👾 CLOSING\n")
				// interrupt fn to make blocking submit successful.
				close(ch)
				wg.Wait()
				select {
				case <-errCh:
					Fail("blocking submit when pool is full should not return error")
				default:
				}
			})
		})
	})
})
