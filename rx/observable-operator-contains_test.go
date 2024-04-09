package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

func two(i rx.Item[int]) bool {
	return i.V == 2
}

var _ = Describe("Observable operator", func() {
	Context("Contains", func() {
		When("sequence contains item", func() {
			It("🧪 should: result in true", func() {
				// rxgo: Test_Observable_Contain
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					testObservable[int](ctx, 1, 2, 3).Contains(two),
					rx.IsTrue[int]{},
				)
			})
		})

		When("sequence does not contain item", func() {
			It("🧪 should: result in false", func() {
				// rxgo: Test_Observable_Contain
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					testObservable[int](ctx, 1, 4, 3).Contains(two),
					rx.IsFalse[int]{},
				)
			})
		})

		Context("Parallel", func() {
			When("sequence contains item", func() {
				It("🧪 should: result in true", func() {
					// rxgo: Test_Observable_Contain_Parallel
					defer leaktest.Check(GinkgoT())()

					/*
						TODO(impl): CPUPool
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx,
							testObservable[int](ctx, 1, 2, 3).Contains(two,
								rx.WithContext[int](ctx),
								rx.WithCPUPool[int](),
							),
							rx.IsTrue[int]{},
						)
					*/
				})
			})

			When("sequence does not contain item", func() {
				It("🧪 should: result in false", func() {
					// rxgo: Test_Observable_Contain_Parallel
					defer leaktest.Check(GinkgoT())()

					/*
						TODO(impl): CPUPool
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						rx.Assert(ctx,
							testObservable[int](ctx, 1, 4, 3).Contains(two,
								rx.WithContext[int](ctx),
								rx.WithCPUPool[int](),
							),
							rx.IsFalse[int]{},
						)
					*/
				})
			})
		})
	})
})
