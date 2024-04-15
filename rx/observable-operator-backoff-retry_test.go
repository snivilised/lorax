package rx_test

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
