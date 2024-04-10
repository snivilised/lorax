package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

func even(item rx.Item[int]) bool {
	return item.V%2 == 0
}

var _ = Describe("Observable operator", func() {
	Context("Filter", func() {
		When("principle", func() {
			It("🧪 should: return filtered items", func() {
				// rxgo: Test_Observable_Filter
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4).Filter(even)
				rx.Assert(ctx, obs,
					rx.HasItems[int]{
						Expected: []int{2, 4},
					},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: return filtered items", func() {
					// rxgo: Test_Observable_Filter_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := testObservable[int](ctx, 1, 2, 3, 4).Filter(even,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs,
						rx.HasItemsNoOrder[int]{
							Expected: []int{2, 4},
						},
						rx.HasNoError[int]{},
					)
				})
			})
		})
	})
})
