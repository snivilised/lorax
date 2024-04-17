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
	"github.com/snivilised/lorax/enums"
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	XContext("Reduce", decorators.Label("broken by reduce acc"), func() {
		When("using Range", func() {
			It("🧪 should: compute reduction ok", func() {
				// rxgo: Test_Observable_Reduce
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range(&rx.NumericRangeIterator[int]{
					StartAt: 1,
					Whilst:  rx.LessThan(10001),
				}).Reduce( // 1, 10000
					func(_ context.Context, acc, num rx.Item[int]) (int, error) {
						return acc.V + num.Num(), nil
					},
				)
				rx.Assert(ctx, obs,
					rx.HasItem[int]{
						Expected: 50005000,
					},
					rx.HasNoError[int]{},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: result in no value", func() {
				// rxgo: Test_Observable_Reduce_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().Reduce(
					func(_ context.Context, _, _ rx.Item[int]) (int, error) {
						panic("apply func should not be called for an Empty iterable")
					},
				)
				rx.Assert(ctx, obs,
					rx.IsEmpty[int]{},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, errFoo, 4, 5).Reduce(
						func(_ context.Context, _, _ rx.Item[int]) (int, error) {
							return 0, nil
						},
					)
					rx.Assert(ctx, obs,
						rx.IsEmpty[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						})
				})
			})
		})

		Context("Parallel", func() {
			When("using Range", func() {
				XIt("🧪 should: compute reduction ok", decorators.Label("repairing"), func() {
					// rxgo: Test_Observable_Reduce_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Range(&rx.NumericRangeIterator[int]{
						StartAt: 1,
						Whilst:  rx.LessThan(6),
					}).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							return acc.Num() + num.Num(), nil
						}, rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasItem[int]{
							Expected: 50005000,
						},
						rx.HasNoError[int]{},
					)
				})
			})
		})

		Context("Parallel/Error", func() {
			When("using Range", func() {
				XIt("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range(&rx.NumericRangeIterator[int]{
						StartAt: 1,
						Whilst:  rx.LessThan(10001),
					}).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							if num.Num() == 1000 {
								return 0, errFoo
							}
							return acc.Num() + num.Num(), nil
						}, rx.WithContext[int](ctx), rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})

			When("error with error strategy", func() {
				XIt("🧪 should: result in error", func() {
					// rxgo: Test_Observable_Reduce_Parallel_WithErrorStrategy
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range(&rx.NumericRangeIterator[int]{
						StartAt: 1,
						Whilst:  rx.LessThan(10001),
					}).Reduce(
						func(_ context.Context, acc, num rx.Item[int]) (int, error) {
							if num.Num() == 1 {
								return 0, errFoo
							}
							return acc.Num() + num.Num(), nil
						}, rx.WithCPUPool[int](), rx.WithErrorStrategy[int](enums.ContinueOnError),
					)
					rx.Assert(ctx, obs,
						rx.HasItem[int]{
							Expected: 50004999,
						},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
