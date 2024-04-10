package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Find", func() {
		Context("not empty", func() {
			When("is present", func() {
				It("🧪 should: return requested item", func() {
					// rxgo: Test_Observable_Find_NotEmpty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := testObservable[int](ctx, 1, 2, 3).Find(func(item rx.Item[int]) bool {
						return item.V == 2
					})
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 2,
					})
				})
			})

			When("is not present", func() {
				It("🧪 should: return nothing", func() {
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := testObservable[int](ctx, 1, 2, 3).Find(func(item rx.Item[int]) bool {
						return item.V == 99
					})
					rx.Assert(ctx, obs, rx.IsEmpty[int]{})
				})
			})
		})

		When("empty", func() {
			It("🧪 should: return nothing", func() {
				// rxgo: Test_Observable_Find_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().Find(func(_ rx.Item[int]) bool {
					return true
				})
				rx.Assert(ctx, obs, rx.IsEmpty[int]{})
			})
		})
	})
})
