package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("ElementAt", func() {
		When("using Range", func() {
			It("🧪 should: extract item at specified index", func() {
				// rxgo: Test_Observable_ElementAt
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range[int](0, 100).ElementAt(99)
				rx.Assert(ctx, obs, rx.HasNumbers[int]{
					Expected: []int{99},
				})
			})
		})

		Context("Errors", func() {
			When("index specified is out of bounds", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_ElementAt_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 0, 1, 2, 3, 4).ElementAt(10)
					rx.Assert(ctx, obs, rx.IsEmpty[int]{}, rx.HasAnError[int]{})
				})
			})
		})

		Context("Parallel", func() {
			When("using Range", func() {
				It("🧪 should: extract item at specified index", func() {
					// rxgo: Test_Observable_ElementAt_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Range[int](0, 100).ElementAt(99, rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.HasNumbers[int]{
						Expected: []int{99},
					})
				})
			})
		})
	})
})
