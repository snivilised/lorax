package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("FlatMap", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_FlatMap
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(item rx.Item[int]) rx.Observable[int] {
					return testObservable[int](ctx, item.V+1, item.V*10)
				})
				rx.Assert[int](ctx, obs, rx.HasItems[int]{
					Expected: []int{2, 10, 3, 20, 4, 30},
				})
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Error1
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.V == 2 {
							return testObservable[int](ctx, errFoo)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs, rx.HasItems[int]{
						Expected: []int{2, 10},
					}, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Error2
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, errFoo, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.IsError() {
							return testObservable[int](ctx, 0)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs,
						rx.HasItems[int]{
							Expected: []int{2, 10, 0, 4, 30},
						}, rx.HasNoError[int]{},
					)
				})
			})
		})

		Context("Parallel", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						return testObservable[int](ctx, i.V+1, i.V*10)
					}, rx.WithCPUPool[int]())
					rx.Assert[int](ctx, obs, rx.HasItemsNoOrder[int]{
						Expected: []int{2, 10, 3, 20, 4, 30},
					})
				})
			})
		})

		Context("Parallel/Error", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_FlatMap_Parallel_Error1
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 3).FlatMap(func(i rx.Item[int]) rx.Observable[int] {
						if i.V == 2 {
							return testObservable[int](ctx, errFoo)
						}
						return testObservable[int](ctx, i.V+1, i.V*10)
					})
					rx.Assert[int](ctx, obs,
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
