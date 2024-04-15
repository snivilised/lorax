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
	Context("FlatMap", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_FlatMap
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(item rx.Item[int]) rx.Observable[int] {
					return testObservable[int](ctx, item.V+1, item.V*10)
				})
				rx.Assert[int](ctx, obs, rx.HasItems[int]{
					Expected: []int{2, 10, 3, 20, 4, 30},
				})
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Error1
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.V == 2 {
							return testObservable[int](ctx, errFoo)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs, rx.HasItems[int]{
						Expected: []int{2, 10},
					}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Error2
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, errFoo, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.IsError() {
							return testObservable[int](ctx, 0)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs,
						rx.HasItems[int]{
							Expected: []int{2, 10, 0, 4, 30},
						}, rx.HasNoError[int]{},
					)
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						return testObservable[int](ctx, i.V+1, i.V*10)
					}, rx.WithCPUPool[int]())
					rx.Assert[int](ctx, obs, rx.HasItemsNoOrder[int]{
						Expected: []int{2, 10, 3, 20, 4, 30},
					})
				})
			})
		})

		Context("Parallel/Error", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Parallel_Error1
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.V == 2 {
							return testObservable[int](ctx, errFoo)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs,
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
