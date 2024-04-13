package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("StartWith", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_StartWithIterable
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 4, 5, 6).StartWith(
					testObservable[int](ctx, 1, 2, 3),
				)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, 3, 4, 5, 6},
					},
					rx.HasNoError[int]{})
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_StartWithIterable_Error1
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 4, 5, 6).StartWith(
						testObservable[int](ctx, 1, errFoo, 3),
					)
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{1},
						},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})

			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_StartWithIterable_Error2
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 4, errFoo, 6).StartWith(
						testObservable[int](ctx, 1, 2, 3))
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{1, 2, 3, 4},
						},
						rx.HasError[int]{
							Expected: []error{errFoo},
						})
				})
			})
		})
	})
})
