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
	Context("ZipFromIterable", func() {
		When("source and other observers same length", func() {
			It("🧪 should: emit zipped elements", func() {
				// rxgo: Test_Observable_ZipFromObservable
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2, 3)
				obs2 := testObservable[int](ctx, 10, 20, 30)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.ContainItems[int]{
					Expected: []int{11, 22, 33},
				})
			})
		})

		When("source observer longer than other", func() {
			It("🧪 should: omit zip from trailing items", func() {
				// rxgo: Test_Observable_ZipFromObservable_DifferentLength1
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2, 3)
				obs2 := testObservable[int](ctx, 10, 20)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.ContainItems[int]{
					Expected: []int{11, 22},
				})
			})
		})

		When("source observer shorter than other", func() {
			It("🧪 should: omit zip from trailing items", func() {
				// rxgo: Test_Observable_ZipFromObservable_DifferentLength2
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2)
				obs2 := testObservable[int](ctx, 10, 20, 30)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.ContainItems[int]{
					Expected: []int{11, 22},
				})
			})
		})
	})
})
