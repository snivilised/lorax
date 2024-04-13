package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Skip", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Skip
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 0, 1, 2, 3, 4, 5).Skip(3)
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{3, 4, 5},
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Skip_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 0, 1, 2, 3, 4, 5).Skip(3,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{3, 4, 5},
						},
					)
				})
			})
		})
	})

	Context("SkipLast", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SkipLast
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 0, 1, 2, 3, 4, 5).SkipLast(3)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{0, 1, 2},
					},
				)
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_SkipLast_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 0, 1, 2, 3, 4, 5).SkipLast(3,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{0, 1, 2},
						},
					)
				})
			})
		})
	})

	Context("SkipWhile", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SkipWhile
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).SkipWhile(
					func(i rx.Item[int]) bool {
						return i.V != 3
					},
				)

				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{3, 4, 5},
				},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_SkipWhile_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3, 4, 5).SkipWhile(
						func(i rx.Item[int]) bool {
							return i.V != 3
						},
						rx.WithCPUPool[int](),
					)

					rx.Assert(ctx, obs, rx.HasItems[int]{
						Expected: []int{3, 4, 5},
					}, rx.HasNoError[int]{})
				})
			})
		})
	})
})
