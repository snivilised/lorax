package ants_test

import (
	"runtime"
	"sync"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ok
	. "github.com/onsi/gomega"    //nolint:revive // ok

	"github.com/snivilised/lorax/internal/ants"
)

var _ = Describe("Ants", func() {
	Context("Submit", func() {
		When("default pool", func() {
			It("🧪 should: run ok", func() {
				// TestAntsPool
				defer leaktest.Check(GinkgoT())()
				defer ants.Release()

				var (
					wg  sync.WaitGroup
					err error
				)

				for i := 0; i < n; i++ {
					wg.Add(1)
					err = ants.Submit(func() {
						demoFunc()
						wg.Done()
					})
				}
				wg.Wait()

				GinkgoWriter.Printf("pool, capacity:%d", ants.Cap())
				GinkgoWriter.Printf("pool, running workers number:%d", ants.Running())
				GinkgoWriter.Printf("pool, free workers number:%d", ants.Free())

				mem := runtime.MemStats{}
				runtime.ReadMemStats(&mem)
				curMem = mem.TotalAlloc/MiB - curMem

				Expect(err).To(Succeed())
				GinkgoWriter.Printf("memory usage:%d MB", curMem)
			})
		})

		When("non-blocking", func() {
			It("🧪 should: not fail", func() {
				// TestNonblockingSubmit
				// ??? defer leaktest.Check(GinkgoT())()

				poolSize := 10
				p, err := ants.NewPool(poolSize, ants.WithNonblocking(true))
				Expect(err).To(Succeed(), "create TimingPool failed")

				defer p.Release()

				for i := 0; i < poolSize-1; i++ {
					Expect(p.Submit(longRunningFunc)).To(Succeed(),
						"nonblocking submit when pool is not full shouldn't return error",
					)
				}
				ch := make(chan struct{})
				ch1 := make(chan struct{})
				f := func() {
					<-ch
					close(ch1)
				}
				// p is full now.
				Expect(p.Submit(f)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)
				Expect(p.Submit(demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
					"nonblocking submit when pool is full should get an ErrPoolOverload",
				)

				// interrupt f to get an available worker
				close(ch)
				<-ch1
				Expect(p.Submit(demoFunc)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)
			})
		})

		When("max blocking", func() {
			It("🧪 should: not fail", func() {
				// TestMaxBlockingSubmit
				// ??? defer leaktest.Check(GinkgoT())()

				poolSize := 10
				p, err := ants.NewPool(poolSize, ants.WithMaxBlockingTasks(1))
				Expect(err).To(Succeed(), "create TimingPool failed")

				defer p.Release()

				for i := 0; i < poolSize-1; i++ {
					Expect(p.Submit(longRunningFunc)).To(Succeed(),
						"blocking submit when pool is not full shouldn't return error",
					)
				}
				ch := make(chan struct{})
				f := func() {
					<-ch
				}
				// p is full now.
				Expect(p.Submit(f)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)

				var wg sync.WaitGroup
				wg.Add(1)
				errCh := make(chan error, 1)
				go func() {
					// should be blocked. blocking num == 1
					if err := p.Submit(demoFunc); err != nil {
						errCh <- err
					}
					wg.Done()
				}()
				time.Sleep(1 * time.Second)
				// already reached max blocking limit
				Expect(p.Submit(demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
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
