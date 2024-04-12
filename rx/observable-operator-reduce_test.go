package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Reduce", func() {
		When("using Range", func() {
			It("🧪 should: compute reduction ok", func() {
				// rxgo: Test_Observable_Reduce
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range[int](1, 10000).Reduce(
					func(_ context.Context, acc, num rx.Item[int]) (int, error) {
						return acc.V + num.N, nil
					},
				)
				rx.Assert(ctx, obs,
					rx.HasItem[int]{
						Expected: 50005000,
					},
					rx.HasNoError[int]{},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: result in no value", func() {
				// rxgo: Test_Observable_Reduce_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().Reduce(
					func(_ context.Context, _, _ rx.Item[int]) (int, error) {
						panic("apply func should not be called for an Empty iterable")
					},
				)
				rx.Assert(ctx, obs,
					rx.IsEmpty[int]{},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, errFoo, 4, 5).Reduce(
						func(_ context.Context, _, _ rx.Item[int]) (int, error) {
							return 0, nil
						},
					)
					rx.Assert(ctx, obs,
						rx.IsEmpty[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						})
				})
			})

			When("return error", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_ReturnError
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservableN[int](ctx, 1, 2, 3).Reduce(
						func(_ context.Context, _, num rx.Item[int]) (int, error) {
							if num.N == 2 {
								return 0, errFoo
							}

							return num.N, nil
						},
					)
					rx.Assert(ctx, obs,
						rx.IsEmpty[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						})
				})
			})
		})

		Context("Parallel", func() {
			When("using Range", func() {
				XIt("🧪 should: compute reduction ok", func() {
					// rxgo: Test_Observable_Reduce_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range[int](1, 10000).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							return acc.N + num.N, nil
						}, rx.WithCPUPool[int]())
					rx.Assert(ctx, obs,
						rx.HasItem[int]{
							Expected: 50005000,
						},
						rx.HasNoError[int]{},
					)
				})
			})
		})

		Context("Parallel/Error", func() {
			When("using Range", func() {
				XIt("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range[int](1, 10000).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							if num.N == 1000 {
								return 0, errFoo
							}
							return acc.N + num.N, nil
						}, rx.WithContext[int](ctx), rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})

			When("error with error strategy", func() {
				XIt("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Parallel_WithErrorStrategy
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range[int](1, 10000).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							if num.N == 1 {
								return 0, errFoo
							}
							return acc.N + num.N, nil
						}, rx.WithCPUPool[int](), rx.WithErrorStrategy[int](rx.ContinueOnError),
					)
					rx.Assert(ctx, obs,
						rx.HasItem[int]{
							Expected: 50004999,
						},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
