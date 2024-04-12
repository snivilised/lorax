package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Scan", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Scan
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4, 5).Scan(
					func(_ context.Context, x, y rx.Item[int]) (int, error) {
						return x.V + y.V, nil
					},
				)
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{1, 3, 6, 10, 15},
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Scan_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3, 4, 5).Scan(
						func(_ context.Context, x, y rx.Item[int]) (int, error) {
							return x.V + y.V, nil
						},
						rx.WithCPUPool[int](),
					)

					rx.Assert(ctx, obs, rx.HasItemsNoOrder[int]{
						Expected: []int{1, 3, 6, 10, 15},
					})
				})
			})
		})
	})
})
