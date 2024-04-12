package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Retry", func() {
		When("principle", func() {
			It("🧪 should: retry", func() {
				// rxgo: Test_Observable_Retry
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				i := 0
				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					if i == 2 {
						next <- rx.Of(3)
					} else {
						i++
						next <- rx.Error[int](errFoo)
					}
				}}).Retry(3, func(_ error) bool {
					return true
				})
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, 1, 2, 1, 2, 3},
					},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Errors", func() {
			When("retry error", func() {
				It("🧪 should: retry", func() {
					// rxgo: Test_Observable_Retry_Error_ShouldRetry
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(1)
						next <- rx.Of(2)
						next <- rx.Error[int](errFoo)
					}}).Retry(3, func(_ error) bool {
						return true
					})
					rx.Assert(ctx, obs, rx.HasItems[int]{
						Expected: []int{1, 2, 1, 2, 1, 2, 1, 2},
					}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("retry error", func() {
				It("🧪 should: not retry", func() {
					// rxgo: Test_Observable_Retry_Error_ShouldNotRetry
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(1)
						next <- rx.Of(2)
						next <- rx.Error[int](errFoo)
					}}).Retry(3, func(_ error) bool {
						return false
					})
					rx.Assert(ctx, obs, rx.HasItems[int]{
						Expected: []int{1, 2},
					}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})
		})
	})
})
