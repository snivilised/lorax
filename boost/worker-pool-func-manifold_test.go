package boost_test

import (
	"context"
	"runtime"
	"sync"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/lorax/internal/ants"
)

func produce(ctx context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg boost.WaitGroup,
) {
	defer wg.Done()

	for i, n := 0, 100; i < n; i++ {
		_ = pool.Post(ctx, Param)
	}

	pool.Conclude(ctx)
}

func inject(ctx context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg boost.WaitGroup,
) {
	defer wg.Done()

	ch := pool.Source(ctx, wg)
	for i, n := 0, 100; i < n; i++ {
		ch <- Param
	}

	close(ch)
}

func consume(_ context.Context,
	pool *boost.ManifoldFuncPool[int, int],
	wg boost.WaitGroup,
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
						ctx, demoPoolManifoldFunc, &wg,
						boost.WithSize(AntsSize),
						boost.WithOutput(10, CheckCloseInterval, TimeoutOnSend),
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
						ctx, demoPoolManifoldFunc, &wg,
						boost.WithSize(AntsSize),
					)

					defer pool.Release(ctx)

					wg.Add(1)
					go produce(ctx, pool, &wg)

					wg.Wait()
					GinkgoWriter.Printf("pool with func, no of running workers:%d\n",
						pool.Running(),
					)
					ShowMemStats()

					Expect(err).To(Succeed())
				})
			})

			Context("with input stream", func() {
				It("🧪 should: not fail", func(specCtx SpecContext) {
					// TestNonblockingSubmit
					var wg sync.WaitGroup

					ctx, cancel := context.WithCancel(specCtx)
					defer cancel()

					pool, err := boost.NewManifoldFuncPool(
						ctx, demoPoolManifoldFunc, &wg,
						boost.WithSize(AntsSize),
						boost.WithInput(InputBufferSize),
						boost.WithOutput(10, CheckCloseInterval, TimeoutOnSend),
					)

					defer pool.Release(ctx)

					wg.Add(1)
					go inject(ctx, pool, &wg)

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

			Context("cancelled", func() {
				Context("without consumer", func() {
					It("🧪 should: not fail", func(specCtx SpecContext) {
						// TestNonblockingSubmit
						var wg sync.WaitGroup

						ctx, cancel := context.WithCancel(specCtx)
						defer cancel()

						pool, err := boost.NewManifoldFuncPool(
							ctx, demoPoolManifoldFunc, &wg,
							boost.WithSize(AntsSize),
						)

						defer pool.Release(ctx)

						wg.Add(1)
						go func(ctx context.Context,
							pool *boost.ManifoldFuncPool[int, int],
							wg boost.WaitGroup,
						) {
							defer wg.Done()

							for i, n := 0, 100; i < n; i++ {
								_ = pool.Post(ctx, Param)

								if i > 10 {
									cancel()
									break
								}
							}
							pool.Conclude(ctx)
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

			Context("timeout on send, with cancellation monitor", func() {
				When("output requested, but accidentally not consumed by client", func() {
					It("🧪 should: cancel context and terminate", func(specCtx SpecContext) {
						// TestNonblockingSubmit
						var wg sync.WaitGroup

						ctx, cancel := context.WithCancel(specCtx)
						defer cancel()

						pool, err := boost.NewManifoldFuncPool(
							ctx, demoPoolManifoldFunc, &wg,
							boost.WithSize(AntsSize),
							boost.WithInput(InputBufferSize),
							boost.WithOutput(10, CheckCloseInterval, TimeoutOnSend),
						)

						defer pool.Release(ctx)

						wg.Add(1)
						go inject(ctx, pool, &wg)

						boost.StartCancellationMonitor(ctx,
							cancel,
							&wg,
							pool.CancelCh(),
							func() {},
						)
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

		Context("IfOption", func() {
			When("true", func() {
				It("🧪 should: use option", func(specCtx SpecContext) {
					ctx, cancel := context.WithCancel(specCtx)
					defer cancel()

					var wg sync.WaitGroup

					const (
						poolSize = 10
					)

					pool, _ := boost.NewManifoldFuncPool(
						ctx, demoPoolManifoldFunc, &wg,
						ants.If(true, ants.WithSize(poolSize)),
						boost.WithInput(InputBufferSize),
						boost.WithOutput(10, CheckCloseInterval, TimeoutOnSend),
					)

					options := pool.GetOptions()
					Expect(options.Size).To(BeEquivalentTo(poolSize))
				})
			})

			When("false", func() {
				It("🧪 should: use option", func(specCtx SpecContext) {
					ctx, cancel := context.WithCancel(specCtx)
					defer cancel()

					var wg sync.WaitGroup

					const (
						poolSize = 10
					)

					pool, _ := boost.NewManifoldFuncPool(
						ctx, demoPoolManifoldFunc, &wg,
						ants.If(false, ants.WithSize(poolSize)),
						boost.WithInput(InputBufferSize),
						boost.WithOutput(10, CheckCloseInterval, TimeoutOnSend),
					)

					options := pool.GetOptions()
					Expect(options.Size).To(BeEquivalentTo(runtime.NumCPU()))
				})
			})
		})
	})
})
