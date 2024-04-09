package rx_test

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Backoff Retry", func() {
		Context("principle", func() {
			It("🧪 should: succeed after retry within max retries", func() {
				// rxgo: Test_Observable_BackOffRetry
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				i := 0
				backOffCfg := backoff.NewExponentialBackOff()
				backOffCfg.InitialInterval = time.Nanosecond
				obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
					next <- rx.Of(1)
					next <- rx.Of(2)
					if i == 2 {
						next <- rx.Of(3)
					} else {
						i++
						next <- rx.Error[int](errFoo)
					}
				}}).BackOffRetry(backoff.WithMaxRetries(backOffCfg, 3))
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, 1, 2, 1, 2, 3},
					},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Errors", func() {
			Context("given: foo", func() {
				It("🧪 should: fail after max retries exceeded", func() {
					// rxgo: Test_Observable_BackOffRetry_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					backOffCfg := backoff.NewExponentialBackOff()
					backOffCfg.InitialInterval = time.Nanosecond
					obs := rx.Defer([]rx.Producer[int]{func(_ context.Context, next chan<- rx.Item[int]) {
						next <- rx.Of(1)
						next <- rx.Of(2)
						next <- rx.Error[int](errFoo)
					}}).BackOffRetry(backoff.WithMaxRetries(backOffCfg, 3))
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{1, 2, 1, 2, 1, 2, 1, 2},
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
