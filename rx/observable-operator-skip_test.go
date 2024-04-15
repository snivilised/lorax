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
