package boost_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/boost"
)

const interval = time.Second / 100

func produce(ctx context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for i, n := 0, 100; i < n; i++ {
		_ = pool.Post(ctx, Param)
	}

	pool.EndWork(ctx, interval)
}

func consume(_ context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for output := range pool.Observe() {
		Expect(output.Error).To(Succeed())
		Expect(output.ID).NotTo(BeEmpty())
		Expect(output.SequenceNo).NotTo(Equal(0))
	}
}

var _ = Describe("WorkerPoolFuncManifold", func() {
	Context("ants", func() {
		When("NonBlocking", func() {
			Context("with consumer", func() {
				It("🧪 should: not fail", func(specCtx SpecContext) {
					// TestNonblockingSubmit
					var wg sync.WaitGroup

					ctx, cancel := context.WithCancel(specCtx)
					defer cancel()

					pool, err := boost.NewManifoldFuncPool(
						ctx, AntsSize, demoPoolManifoldFunc, &wg,
						boost.WithOutput(10),
					)

					defer pool.Release(ctx)

					wg.Add(1)
					go produce(ctx, pool, &wg)

					wg.Add(1)
					go consume(ctx, pool, &wg)

					wg.Wait()
					GinkgoWriter.Printf("pool with func, no of running workers:%d\n",
						pool.Running(),
					)
					ShowMemStats()

					Expect(err).To(Succeed())
				})
			})

			Context("without consumer", func() {
				It("🧪 should: not fail", func(specCtx SpecContext) {
					// TestNonblockingSubmit
					var wg sync.WaitGroup

					ctx, cancel := context.WithCancel(specCtx)
					defer cancel()

					pool, err := boost.NewManifoldFuncPool(
						ctx, AntsSize, demoPoolManifoldFunc, &wg,
					)

					defer pool.Release(ctx)

					wg.Add(1)
					go produce(ctx, pool, &wg)

					// tbd, create a consumer for the result channel
					wg.Wait()
					GinkgoWriter.Printf("pool with func, no of running workers:%d\n",
						pool.Running(),
					)
					ShowMemStats()

					Expect(err).To(Succeed())
				})
			})

			Context("cancelled", func() {
				Context("without consumer", func() {
					It("🧪 should: not fail", func(specCtx SpecContext) {
						// TestNonblockingSubmit
						var wg sync.WaitGroup

						ctx, cancel := context.WithCancel(specCtx)
						defer cancel()

						pool, err := boost.NewManifoldFuncPool(
							ctx, AntsSize, demoPoolManifoldFunc, &wg,
						)

						defer pool.Release(ctx)

						wg.Add(1)
						go func(ctx context.Context,
							pool *boost.ManifoldFuncPool[int, int],
							wg *sync.WaitGroup,
						) {
							defer wg.Done()

							for i, n := 0, 100; i < n; i++ {
								_ = pool.Post(ctx, Param)

								if i > 10 {
									cancel()
									break
								}
							}
							pool.EndWork(ctx, interval)
						}(ctx, pool, &wg)

						wg.Wait()
						GinkgoWriter.Printf("pool with func, no of running workers:%d\n",
							pool.Running(),
						)
						ShowMemStats()

						Expect(err).To(Succeed())
					})
				})
			})
		})
	})
})
