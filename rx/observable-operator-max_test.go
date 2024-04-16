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
	Context("Max", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Max
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range[int](0, 100).Max(
					rx.NumericItemLimitComparator,
					rx.MaxNItemInitLimitInt,
				)
				rx.Assert(ctx, obs, rx.HasNumber[int]{
					Expected: 99,
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Max_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range[int](0, 100).Max(
						rx.NumericItemLimitComparator,
						rx.MaxNItemInitLimitInt,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.HasNumber[int]{
						Expected: 99,
					})
				})
			})
		})
	})
})
