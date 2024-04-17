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
	"github.com/onsi/ginkgo/v2/dsl/decorators"

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Average", func() {
		Context("principle", func() {
			It("🧪 should: calculate the average", func() {
				// rxgo: Test_Observable_AverageFloat32
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					testObservable[float32](ctx,
						float32(1), float32(20),
					).Average(rx.Calc[float32]()),
					rx.HasItem[float32]{
						Expected: 10.5,
					},
				)
				rx.Assert(ctx, testObservable[float32](ctx,
					float32(1), float32(20),
				).Average(rx.Calc[float32]()),
					rx.HasItem[float32]{
						Expected: 10.5,
					},
				)
			})
		})

		Context("Empty", func() {
			Context("given: type float32", func() {
				It("🧪 should: calculate the average as 0", func() {
					// rxgo: Test_Observable_AverageFloat32_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx,
						rx.Empty[float32]().Average(rx.Calc[float32]()),
						rx.HasItem[float32]{
							Expected: 0.0,
						},
					)
				})
			})
		})

		Context("Errors", func() {
			Context("given: type float32", func() {
				It("🧪 should: contain error", func() {
					// rxgo: Test_Observable_AverageFloat32_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						"foo",
					).Average(rx.Calc[float32]()),
						rx.HasAnError[float32]{},
					)
				})
			})
		})

		Context("Parallel", func() {
			Context("given: type float32", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_AverageFloat32_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						float32(1), float32(20),
					).Average(rx.Calc[float32]()),
						rx.HasItem[float32]{
							Expected: float32(10.5),
						},
					)

					rx.Assert(ctx, testObservable[float32](ctx,
						float32(1), float32(20),
					).Average(rx.Calc[float32]()),
						rx.HasItem[float32]{
							Expected: float32(10.5),
						},
					)
				})
			})
		})

		Context("Parallel/Error", func() {
			Context("given: foo", func() {
				XIt("🧪 should: ", decorators.Label("broken average.gatherNext"), func() {
					// rxgo: Test_Observable_AverageFloat32_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						"foo",
					).Average(rx.Calc[float32](),
						rx.WithContext[float32](ctx), rx.WithCPUPool[float32](),
					),
						rx.HasAnError[float32]{},
					)
				})
			})
		})
	})
})
