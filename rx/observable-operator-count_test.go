package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Count", func() {
		Context("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Count
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx, rx.Range[int](1, 100).Count(),
					rx.HasNumber[int]{
						Expected: 100,
					})
			})
		})

		Context("Parallel", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Count_Parallel
					defer leaktest.Check(GinkgoT())()

					/*
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx, rx.Range[int](1, 100).Count(
							rx.WithCPUPool[int](),
						),
							rx.HasNumber[int]{
								Expected: 100,
							},
						)
					*/
				})
			})
		})
	})
})
