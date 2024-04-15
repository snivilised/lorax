package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var positiveN = func(it rx.Item[int]) bool {
	return it.Num() > 0
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

					rx.Assert(ctx, rx.Range[int](1, 3).All(positiveN),
						rx.IsTrue[int]{},
						rx.HasNoError[int]{},
					)
				})

				Context("HasTrue", func() {
					It("🧪 should: return true", func() {
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, rx.Range[int](1, 3).All(positiveN),
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

		XContext("Parallel", func() {
			Context("all true", func() {
				Context("given: foo", func() {
					It("🧪 should: ", func() {
						// rxgo: Test_Observable_All_Parallel_True
						defer leaktest.Check(GinkgoT())()

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, rx.Range[int](1, 3).All(positiveN,
							rx.WithContext[int](ctx),
							rx.WithCPUPool[int](), // not supported yet
						),
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
							rx.WithCPUPool[int](), // not supported yet
						),
							rx.IsFalse[int]{},
							rx.HasNoError[int]{},
						)
					})
				})
			})
		})

		XContext("Error", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_All_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[int](ctx, 1, errFoo, 3).All(negative,
						rx.WithContext[int](ctx),
						rx.WithCPUPool[int](), // not supported yet
					),
						rx.IsFalse[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
