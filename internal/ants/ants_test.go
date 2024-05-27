package ants_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ok
	. "github.com/onsi/gomega"    //nolint:revive // ok

	"github.com/snivilised/lorax/internal/ants"
)

var _ = Describe("Ants", func() {
	Context("NewPool", func() {
		Context("Submit", func() {
			When("non-blocking", func() {
				It("🧪 should: not fail", func() {
					// TestNonblockingSubmit
					// ??? defer leaktest.Check(GinkgoT())()

					poolSize := 10
					pool, err := ants.NewPool(poolSize, ants.WithNonblocking(true))
					Expect(err).To(Succeed(), "create TimingPool failed")

					defer pool.Release()

					for i := 0; i < poolSize-1; i++ {
						Expect(pool.Submit(longRunningFunc)).To(Succeed(),
							"nonblocking submit when pool is not full shouldn't return error",
						)
					}
					firstCh := make(chan struct{})
					secondCh := make(chan struct{})
					fn := func() {
						<-firstCh
						close(secondCh)
					}
					// p is full now.
					Expect(pool.Submit(fn)).To(Succeed(),
						"nonblocking submit when pool is not full shouldn't return error",
					)
					Expect(pool.Submit(demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
						"nonblocking submit when pool is full should get an ErrPoolOverload",
					)

					// interrupt fn to get an available worker
					close(firstCh)
					<-secondCh
					Expect(pool.Submit(demoFunc)).To(Succeed(),
						"nonblocking submit when pool is not full shouldn't return error",
					)
				})
			})

			When("max blocking", func() {
				It("🧪 should: not fail", func() {
					// TestMaxBlockingSubmit
					// ??? defer leaktest.Check(GinkgoT())()

					poolSize := 10
					pool, err := ants.NewPool(poolSize, ants.WithMaxBlockingTasks(1))
					Expect(err).To(Succeed(), "create TimingPool failed")

					defer pool.Release()

					for i := 0; i < poolSize-1; i++ {
						Expect(pool.Submit(longRunningFunc)).To(Succeed(),
							"blocking submit when pool is not full shouldn't return error",
						)
					}
					ch := make(chan struct{})
					fn := func() {
						<-ch
					}
					// p is full now.
					Expect(pool.Submit(fn)).To(Succeed(),
						"nonblocking submit when pool is not full shouldn't return error",
					)

					var wg sync.WaitGroup
					wg.Add(1)
					errCh := make(chan error, 1)
					go func() {
						// should be blocked. blocking num == 1
						if err := pool.Submit(demoFunc); err != nil {
							errCh <- err
						}
						wg.Done()
					}()
					time.Sleep(1 * time.Second)
					// already reached max blocking limit
					Expect(pool.Submit(demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
						"blocking submit when pool reach max blocking submit should return ErrPoolOverload",
					)

					// interrupt f to make blocking submit successful.
					close(ch)
					wg.Wait()
					select {
					case <-errCh:
						// t.Fatalf("blocking submit when pool is full should not return error")
						Fail("blocking submit when pool is full should not return error")
					default:
					}
				})
			})
		})
	})

	Context("NewPoolWithFunc", func() {
		Context("Invoke", func() {
			When("waiting to get worker", func() {
				It("🧪 should: not fail", func() {
					// TestAntsPoolWithFuncWaitToGetWorker

					var wg sync.WaitGroup
					pool, _ := ants.NewPoolWithFunc(AntsSize, func(i ants.InputParam) {
						demoPoolFunc(i)
						wg.Done()
					})
					defer pool.Release()

					for i := 0; i < n; i++ {
						wg.Add(1)
						_ = pool.Invoke(Param)
					}
					wg.Wait()
					GinkgoWriter.Printf("pool with func, running workers number:%d\n",
						pool.Running(),
					)
					ShowMemStats()
				})
			})

			When("waiting to get worker with pre malloc", func() {
				It("🧪 should: not fail", func() {
					// TestAntsPoolWithFuncWaitToGetWorkerPreMalloc

					var wg sync.WaitGroup
					pool, _ := ants.NewPoolWithFunc(AntsSize, func(i ants.InputParam) {
						demoPoolFunc(i)
						wg.Done()
					}, ants.WithPreAlloc(true))
					defer pool.Release()

					for i := 0; i < n; i++ {
						wg.Add(1)
						_ = pool.Invoke(Param)
					}
					wg.Wait()
					GinkgoWriter.Printf("pool with func, running workers number:%d\n",
						pool.Running(),
					)
					ShowMemStats()
				})
			})
		})
	})
})
