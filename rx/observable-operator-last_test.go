package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Last", func() {
		When("not empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Last_NotEmpty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).Last()
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 3,
				})
			})
		})

		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Last_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().Last()
				rx.Assert(ctx, obs, rx.IsEmpty[int]{})
			})
		})

		Context("Parallel", func() {
			When("not empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Last_Parallel_NotEmpty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).Last(
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 3,
					})
				})
			})

			When("empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Last_Parallel_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Empty[int]().Last(rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.IsEmpty[int]{})
				})
			})
		})
	})

	Context("LastOrDefault", func() {
		When("not empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_LastOrDefault_NotEmpty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).LastOrDefault(10)
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 3,
				})
			})
		})

		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_LastOrDefault_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().LastOrDefault(10)
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 10,
				})
			})
		})

		Context("Parallel", func() {
			When("not empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_LastOrDefault_Parallel_NotEmpty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).LastOrDefault(10,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 3,
					})
				})
			})

			When("empty", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_LastOrDefault_Parallel_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Empty[int]().LastOrDefault(10,
						rx.WithCPUPool[int](),
					)
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 10,
					})
				})
			})
		})
	})
})
