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

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Take", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Take
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).Take(3)
				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{1, 2, 3},
					},
				)
			})
		})

		When("interval", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Take_Interval
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Interval(
					rx.WithDuration(time.Nanosecond),
					rx.WithContext[int](ctx),
				).Take(3)

				rx.Assert(ctx, obs,
					rx.HasTickValueCount[int]{
						Expected: 3,
					},
				)
			})
		})
	})

	Context("TakeLast", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_TakeLast
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).TakeLast(3)
				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{3, 4, 5},
					},
				)
			})
		})

		When("principle", func() {
			It("🧪 should: less than nth", func() {
				// rxgo: Test_Observable_TakeLast_LessThanNth
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 4, 5).TakeLast(3)
				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{4, 5},
					},
				)
			})
		})

		When("principle", func() {
			It("🧪 should: less than nth 2", func() {
				// rxgo: Test_Observable_TakeLast_LessThanNth2
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 4, 5).TakeLast(100000)
				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{4, 5},
					},
				)
			})
		})
	})

	Context("TakeUntil", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_TakeUntil
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).TakeUntil(
					func(item rx.Item[int]) bool {
						return item.V == 3
					})
				rx.Assert(ctx, obs, rx.ContainItems[int]{
					Expected: []int{1, 2, 3},
				})
			})
		})
	})

	Context("TakeWhile", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_TakeWhile
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).TakeWhile(
					func(item rx.Item[int]) bool {
						return item.V != 3
					},
				)
				rx.Assert(ctx, obs, rx.ContainItems[int]{
					Expected: []int{1, 2},
				})
			})
		})
	})
})
