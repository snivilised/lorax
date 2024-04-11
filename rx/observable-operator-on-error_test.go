package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("OnErrorResumeNext", func() {
		When("error occurs", func() {
			It("🧪 should: continue with next iterable", func() {
				// rxgo: Test_Observable_OnErrorResumeNext
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				obs := testObservable[int](ctx, 1, 2, errFoo, 4).OnErrorResumeNext(
					func(_ error) rx.Observable[int] {
						return testObservable[int](ctx, 10, 20)
					},
				)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, 10, 20},
					},
					rx.HasNoError[int]{},
				)
			})
		})
	})

	Context("OnErrorReturn", func() {
		When("error occurs", func() {
			It("🧪 should: emit translated error and continue", func() {
				// rxgo: Test_Observable_OnErrorReturn
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, errFoo, 4, errBar, 6).OnErrorReturn(
					func(_ error) int {
						return -1
					},
				)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, -1, 4, -1, 6},
					},
					rx.HasNoError[int]{},
				)
			})
		})
	})

	Context("OnErrorReturnItem", func() {
		When("error occurs", func() {
			It("🧪 should: emit translated error and continue", func() {
				// rxgo: Test_Observable_OnErrorReturnItem
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, errFoo, 4, errBar, 6).OnErrorReturnItem(-1)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{1, 2, -1, 4, -1, 6},
					},
					rx.HasNoError[int]{},
				)
			})
		})
	})
})
