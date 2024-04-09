package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("DefaultIfEmpty", func() {
		When("empty", func() {
			It("🧪 should: return default value", func() {
				// rxgo: Test_Observable_DefaultIfEmpty_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().DefaultIfEmpty(3)
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{3},
				})
			})
		})

		When("not empty", func() {
			It("🧪 should: have emitted values", func() {
				// rxgo: Test_Observable_DefaultIfEmpty_NotEmpty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2).DefaultIfEmpty(3)
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{1, 2},
				})
			})
		})

		Context("Parallel", func() {
			When("empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DefaultIfEmpty_Parallel_Empty
					defer leaktest.Check(GinkgoT())()

					/*
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						obs := rx.Empty[int]().DefaultIfEmpty(3,
							rx.WithCPUPool[int](),
						)
						rx.Assert(ctx, obs, rx.HasItems[int]{
							Expected: []int{3},
						})
					*/
				})
			})

			When("not empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DefaultIfEmpty_Parallel_NotEmpty
					defer leaktest.Check(GinkgoT())()

					/*
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						obs := testObservable[int](ctx, 1, 2).DefaultIfEmpty(3,
							rx.WithCPUPool[int](),
						)
						rx.Assert(ctx, obs, rx.HasItems[int]{
							Expected: []int{1, 2},
						})
					*/
				})
			})
		})

		Context("Parallel/Error", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_
					defer leaktest.Check(GinkgoT())()
				})
			})
		})
	})
})
