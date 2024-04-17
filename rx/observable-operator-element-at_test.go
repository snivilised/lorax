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
	Context("ElementAt", func() {
		When("using Range", func() {
			It("🧪 should: extract item at specified index", func() {
				// rxgo: Test_Observable_ElementAt
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range(&rx.NumericRangeIterator[int]{
					StartAt: 0,
					Whilst:  rx.LessThan(100),
				}).ElementAt(99)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{99},
					},
				)
			})
		})

		Context("Errors", func() {
			When("index specified is out of bounds", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_ElementAt_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 0, 1, 2, 3, 4).ElementAt(10)
					rx.Assert(ctx, obs,
						rx.IsEmpty[int]{},
						rx.HasAnError[int]{},
					)
				})
			})
		})

		Context("Parallel", func() {
			When("using Range", func() {
				It("🧪 should: extract item at specified index", func() {
					// rxgo: Test_Observable_ElementAt_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Range(&rx.NumericRangeIterator[int]{
						StartAt: 0,
						Whilst:  rx.LessThan(11),
					}).ElementAt(10, rx.WithCPUPool[int]())
					rx.Assert(ctx, obs,
						rx.HasItems[int]{
							Expected: []int{10},
						},
					)
				})
			})
		})
	})
})
