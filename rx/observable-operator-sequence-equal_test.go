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
	Context("SequenceEqual", func() {
		When("even sequence", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_EvenSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequence := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213)
				result := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequence,
					rx.NativeItemLimitComparator,
				)
				rx.Assert(ctx, result,
					rx.IsTrue[int]{},
				)
			})
		})

		When("Uneven sequence", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_UnevenSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequence := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213)
				result := testObservable[int](ctx, 2, 5, 12, 43, 15, 100, 213).SequenceEqual(
					sequence,
					rx.NativeItemLimitComparator,
					rx.WithContext[int](ctx),
				)
				rx.Assert(ctx, result,
					rx.IsFalse[int]{},
				)
			})
		})

		When("Different sequence length", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_DifferentLengthSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequenceShorter := testObservable[int](ctx, 2, 5, 12, 43, 98, 100)
				sequenceLonger := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213, 512)

				resultForShorter := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequenceShorter,
					rx.NativeItemLimitComparator,
				)
				rx.Assert(ctx, resultForShorter,
					rx.IsFalse[int]{},
				)

				resultForLonger := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequenceLonger,
					rx.NativeItemLimitComparator,
				)
				rx.Assert(ctx, resultForLonger,
					rx.IsFalse[int]{},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				result := rx.Empty[int]().SequenceEqual(
					rx.Empty[int](),
					rx.NativeItemLimitComparator,
				)
				rx.Assert(ctx, result,
					rx.IsTrue[int]{},
				)
			})
		})
	})
})
