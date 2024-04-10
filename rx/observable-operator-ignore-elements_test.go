package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("IgnoreElements", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_IgnoreElements
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).IgnoreElements()
				rx.Assert(ctx, obs, rx.IsEmpty[int]{})
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_IgnoreElements_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, errFoo, 3).IgnoreElements()
					rx.Assert(ctx, obs,
						rx.IsEmpty[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						})
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_IgnoreElements_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).IgnoreElements(
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.IsEmpty[int]{})
				})
			})
		})

		Context("Parallel/Error", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_IgnoreElements_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, errFoo, 3).IgnoreElements(
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.IsEmpty[int]{}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})
		})
	})
})
