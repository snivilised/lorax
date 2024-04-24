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
	Context("Sum", func() {
		When("principle", func() {
			It("🧪 should: return sum", func() {
				// rxgo: Test_Observable_SumFloat32_OnlyFloat32
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx, testObservable[float32](ctx,
					// explicit cast to float32 is required to match
					// the generic type to ensure observable sends the
					// item. Failure to perform this task results in
					// an error ("channel value: '1' not sent (wrong type?)").
					//
					float32(1.0), float32(2.0), float32(3.0),
				).Sum(rx.WithCalc(rx.Calc[float32]())),
					rx.HasItem[float32]{
						Expected: 6.0,
					},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: return empty result", func() {
				// rxgo: Test_Observable_SumFloat32_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					rx.Empty[float32]().Sum(rx.WithCalc(rx.Calc[float32]())),
					rx.IsEmpty[float32]{},
				)
			})
		})

		// NB(Test_Observable_SumFloat32_DifferentTypes): Sum does
		// not support different types, only values of type T are
		// supported.

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_SumFloat32_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						// we omit the explicit cast to float32 as normally would
						// be the case, to enforce a resulting error.
						//
						1.1, 2.2, 3.3,
					).Sum(rx.WithCalc(rx.Calc[float32]())),
						rx.HasAnError[float32]{},
					)
				})
			})

			Context("missing calc", func() {
				It("🧪 should: raise error", func() {
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						float32(1.0), float32(2.0), float32(3.0),
					).Sum(
					// forget to provide a calculator
					),
						rx.HasError[float32]{
							Expected: []error{rx.MissingCalcError{}},
						},
					)
				})
			})
		})
	})
})
