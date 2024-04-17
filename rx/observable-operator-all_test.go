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

var positive = func(it rx.Item[int]) bool {
	return it.V > 0
}

var negative = func(it rx.Item[int]) bool {
	return it.V < 0
}

var _ = Describe("Observable operator", func() {
	Context("All", func() {
		Context("principle", func() {
			Context("all true", func() {
				It("🧪 should: return true", func() {
					// rxgo: Test_Observable_All_True
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Range(&rx.NumericRangeIterator[int]{
						StartAt: 1,
						Whilst:  rx.LessThan(4),
					}).All(positive)

					rx.Assert(ctx, obs,
						rx.IsTrue[int]{},
						rx.HasNoError[int]{},
					)
				})

				Context("HasTrue", func() {
					It("🧪 should: return true", func() {
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						obs := rx.Range(&rx.NumericRangeIterator[int]{
							StartAt: 1,
							Whilst:  rx.LessThan(4),
						}).All(positive)

						rx.Assert(ctx,
							obs,
							rx.HasTrue[int]{},
						)
					})
				})
			})

			Context("all false", func() {
				It("🧪 should: return false", func() {
					// rxgo: Test_Observable_All_False
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[int](ctx, 1, -2, 3).All(negative),
						rx.IsFalse[int]{},
						rx.HasNoError[int]{},
					)

					// 💥 Warning, it might seem natural to implement this in exactly the opposite
					// way around to the "🧪 should: return true" case, ie using the Range operator, instead
					// of the testObservable. But when the predicate returns false, the internal
					// pipeline is terminated early as there is no need to check further items if
					// a false exists because we just found one. This results in a different execution
					// path that requires different handling by the client. If we attempt to use
					// the Range operator, instead of the testObservable observable, then we end
					// up leaking a Go routine as reported by leaktest:
					//
					// 				Summarizing 1 Failure:
					// [FAIL] Observable operator All principle all false [It] 🧪 should: return false
					// /Users/plastikfan/go/pkg/mod/github.com/fortytw2/leaktest@v1.3.0/leaktest.go:132
					//
					// TODO: verify exactly the reason why this leaktest occurs, for now just
					// take as read.
				})

				Context("HasFalse", func() {
					It("🧪 should: return false", func() {
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, testObservable[int](ctx, -1, 2, 3).All(negative),
							rx.HasFalse[int]{},
						)
					})
				})
			})
		})

		Context("Parallel", func() {
			Context("all true", func() {
				Context("given: foo", func() {
					It("🧪 should: ", func() {
						// rxgo: Test_Observable_All_Parallel_True
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						obs := rx.Range(&rx.NumericRangeIterator[int]{
							StartAt: 1,
							Whilst:  rx.LessThan(4),
						}).All(positive,
							rx.WithContext[int](ctx),
							rx.WithCPUPool[int](),
						)

						rx.Assert(ctx, obs,
							rx.HasTrue[int]{},
							rx.HasNoError[int]{},
						)
					})
				})
			})

			Context("all false", func() {
				Context("given: foo", func() {
					It("🧪 should: ", func() {
						// rxgo: Test_Observable_All_Parallel_False
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, testObservable[int](ctx, 1, -2, 3).All(negative,
							rx.WithContext[int](ctx),
							rx.WithCPUPool[int](),
						),
							rx.IsFalse[int]{},
							rx.HasNoError[int]{},
						)
					})
				})
			})
		})

		Context("Error", func() {
			Context("given: error occurs after predicate failure", func() {
				It("🧪 should: not return error", func() {
					// rxgo: Test_Observable_All_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[int](ctx, 1, errFoo, 3).All(negative,
						rx.WithContext[int](ctx),
						rx.WithCPUPool[int](),
					),
						rx.IsFalse[int]{},
					)
				})
			})

			Context("given: error occurs before predicate failure", func() {
				XIt("🧪 should: return error", decorators.Label("Flakey"), func() {
					// rxgo: Test_Observable_All_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					// Flakey: (possible race)

					// NB: Once an error has occurred, you can't rely on the
					// result of All to be accurate. The result will only
					// reflect the state of processing at the time the error
					// occurred. In this particular case, the result of All will
					// be true and so will rx.IsTrue[int], that is because
					// 1 and 3 were not seen.
					//
					rx.Assert(ctx, testObservable[int](ctx, errFoo, 1, 3).All(negative,
						rx.WithContext[int](ctx),
						rx.WithCPUPool[int](),
					),
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
