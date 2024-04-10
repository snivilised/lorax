package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("First", func() {
		When("not empty", func() {
			It("🧪 should: return first item", func() {
				// rxgo: Test_Observable_First_NotEmpty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).First()
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 1,
				})
			})
		})

		When("empty", func() {
			It("🧪 should: return nothing", func() {
				// rxgo: Test_Observable_First_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().First()
				rx.Assert(ctx, obs, rx.IsEmpty[int]{})
			})
		})

		Context("Parallel", func() {
			When("not empty", func() {
				It("🧪 should: return first item", func() {
					// rxgo: Test_Observable_First_Parallel_NotEmpty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).First(rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 1,
					})
				})
			})

			When("empty", func() {
				It("🧪 should: return nothing", func() {
					// rxgo: Test_Observable_First_Parallel_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Empty[int]().First(rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.IsEmpty[int]{})
				})
			})
		})
	})

	Context("FirstOrDefault", func() {
		When("not empty", func() {
			It("🧪 should: return first item", func() {
				// rxgo: Test_Observable_First_NotEmpty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).FirstOrDefault(10)
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 1,
				})
			})
		})

		When("empty", func() {
			It("🧪 should: return default", func() {
				// rxgo: Test_Observable_First_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := rx.Empty[int]().FirstOrDefault(10)
				rx.Assert(ctx, obs, rx.HasItem[int]{
					Expected: 10,
				})
			})
		})

		Context("Parallel", func() {
			When("not empty", func() {
				It("🧪 should: return first item", func() {
					// rxgo: Test_Observable_First_Parallel_NotEmpty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FirstOrDefault(10, rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 1,
					})
				})
			})

			When("empty", func() {
				It("🧪 should: return default ", func() {
					// rxgo: Test_Observable_First_Parallel_Empty
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Empty[int]().FirstOrDefault(10, rx.WithCPUPool[int]())
					rx.Assert(ctx, obs, rx.HasItem[int]{
						Expected: 10,
					})
				})
			})
		})
	})
})
