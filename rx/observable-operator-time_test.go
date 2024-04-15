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
	"fmt"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("TimeInterval", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_TimeInterval
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).TimeInterval()

				rx.Assert(ctx, obs, rx.CustomPredicate[int]{
					Expected: func(actual rx.AssertResources[int]) error {
						items := actual.Opaques()

						if len(items) != 3 {
							return fmt.Errorf("expected 3 items, got %d items", len(items))
						}

						return nil
					},
				})
			})
		})
	})

	Context("Timestamp", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Timestamp
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				observe := testObservable[int](ctx, 1, 2, 3).Timestamp().Observe()

				v, _ := (<-observe).Opaque().(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(1))

				v, _ = (<-observe).Opaque().(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(2))

				v, _ = (<-observe).Opaque().(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(3))
			})
		})
	})
})
