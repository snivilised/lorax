package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Average", func() {
		Context("principle", func() {
			It("🧪 should: calculate the average", func() {
				// rxgo: Test_Observable_AverageFloat32
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					testObservable[float32](ctx,
						float32(1), float32(20),
					).Average(rx.Calc[float32]()),
					rx.HasItem[float32]{
						Expected: 10.5,
					},
				)
				rx.Assert(ctx, testObservable[float32](ctx,
					float32(1), float32(20),
				).Average(rx.Calc[float32]()),
					rx.HasItem[float32]{
						Expected: 10.5,
					},
				)
			})
		})

		Context("Empty", func() {
			Context("given: type float32", func() {
				It("🧪 should: calculate the average as 0", func() {
					// rxgo: Test_Observable_AverageFloat32_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx,
						rx.Empty[float32]().Average(rx.Calc[float32]()),
						rx.HasItem[float32]{
							Expected: 0.0,
						},
					)
				})
			})
		})

		Context("Errors", func() {
			Context("given: type float32", func() {
				It("🧪 should: contain error", func() {
					// rxgo: Test_Observable_AverageFloat32_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						"foo",
					).Average(rx.Calc[float32]()),
						rx.HasAnError[float32]{},
					)
				})
			})
		})

		Context("Parallel", func() {
			Context("given: type float32", func() {
				It("🧪 should: ", func() {
					// rxgo(tbd - parallel not impl yet): Test_Observable_AverageFloat32_Parallel
					defer leaktest.Check(GinkgoT())()

					/*
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, testObservable[float32](ctx,
							float32(1), float32(20),
						).Average(rx.NewCalc[float32]()),
							rx.HasItem[float32]{
								Expected: float32(10.5),
							},
						)

						rx.Assert(ctx, testObservable[float32](ctx,
							float32(1), float32(20),
						).Average(rx.NewCalc[float32]()),
							rx.HasItem[float32]{
								Expected: float32(10.5),
							},
						)
					*/
				})
			})
		})

		Context("Parallel/Error", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo(tbd - parallel not impl yet): Test_Observable_AverageFloat32_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					/*
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, testObservable[float32](ctx,
							"foo",
						).Average(rx.NewCalc[float32](),
							rx.WithContext[float32](ctx), rx.WithCPUPool[float32](),
						),
							rx.HasAnError[float32]{},
						)
					*/
				})
			})
		})
	})
})
