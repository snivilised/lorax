package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Map", func() {
		When("one", func() {
			It("🧪 should: translate all values", func() {
				// rxgo: Test_Observable_Map_One
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).Map(func(_ context.Context, v int) (int, error) {
					return v + 1, nil
				})
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{2, 3, 4},
				},
					rx.HasNoError[int]{},
				)
			})
		})

		When("multiple", func() {
			It("🧪 should: transform all values through all stages", func() {
				// rxgo: Test_Observable_Map_Multiple
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).Map(func(_ context.Context, v int) (int, error) {
					return v + 1, nil
				}).Map(func(_ context.Context, v int) (int, error) {
					return v * 10, nil
				})
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{20, 30, 40},
				}, rx.HasNoError[int]{})
			})
		})

		When("cancel", func() {
			It("🧪 should: result in empty result due to cancellation", func() {
				// rxgo: Test_Observable_Map_Cancel
				defer leaktest.Check(GinkgoT())()

				next := make(chan rx.Item[int])

				ctx, cancel := context.WithCancel(context.Background())
				obs := rx.FromChannel(next).Map(func(_ context.Context, v int) (int, error) {
					return v + 1, nil
				}, rx.WithContext[int](ctx))
				cancel()
				rx.Assert(ctx, obs, rx.IsEmpty[int]{}, rx.HasNoError[int]{})
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: transform valid values and contain error", func() {
					// rxgo: Test_Observable_Map_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3, errFoo).Map(func(_ context.Context, v int) (int, error) {
						return v + 1, nil
					})
					rx.Assert(ctx, obs, rx.HasItems[int]{
						Expected: []int{2, 3, 4},
					}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("value and error", func() {
				It("🧪 should: not transform value that returns error", func() {
					// rxgo: Test_Observable_Map_ReturnValueAndError
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1).Map(func(_ context.Context, _ int) (int, error) {
						return 2, errFoo
					})
					rx.Assert(ctx, obs, rx.IsEmpty[int]{}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("multiple errors", func() {
				It("🧪 should: contain error", func() {
					// rxgo: Test_Observable_Map_Multiple_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					called := false
					obs := testObservable[int](ctx, 1, 2, 3).Map(func(_ context.Context, _ int) (int, error) {
						return 0, errFoo
					}).Map(func(_ context.Context, _ int) (int, error) {
						called = true

						return 0, nil
					})
					rx.Assert(ctx, obs, rx.IsEmpty[int]{}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
					Expect(called).To(BeFalse())
				})
			})
		})

		Context("Parallel", func() {
			It("🧪 should: translate all values", func() {
				// rxgo: Test_Observable_Map_Parallel
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				const length = 10

				ch := make(chan rx.Item[int], length)
				go func() {
					for i := 0; i < length; i++ {
						ch <- rx.Of(i)
					}
					close(ch)
				}()

				obs := rx.FromChannel(ch).Map(func(_ context.Context, v int) (int, error) {
					return v + 1, nil
				}, rx.WithPool[int](length))

				rx.Assert(ctx, obs, rx.HasItemsNoOrder[int]{
					Expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				}, rx.HasNoError[int]{})
			})
		})
	})
})
