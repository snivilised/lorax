package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Max", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Max
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Range[int](0, 100).Max(
					rx.NumericItemLimitComparator,
					rx.MaxNItemInitLimitInt,
				)
				rx.Assert(ctx, obs, rx.HasNumber[int]{
					Expected: 99,
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				XIt("🧪 should: ", func() {
					// rxgo: Test_Observable_Max_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					obs := rx.Range[int](0, 100).Max(
						rx.NumericItemLimitComparator,
						rx.MaxNItemInitLimitInt,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 99,
					})
				})
			})
		})
	})
})