package rx_test

import (
	"context"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("WindowWithCount", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_WindowWithCount
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				observe := testObservable[int](ctx, 1, 2, 3, 4, 5).WindowWithCount(2).Observe()

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItems[int]{
						Expected: []int{1, 2},
					},
				)

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItems[int]{
						Expected: []int{3, 4},
					},
				)

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItem[int]{
						Expected: 5,
					},
				)
			})
		})

		When("Zero count", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_WindowWithCount_ZeroCount
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				observe := testObservable[int](ctx, 1, 2, 3, 4, 5).WindowWithCount(0).Observe()

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItems[int]{
						Expected: []int{1, 2, 3, 4, 5},
					},
				)
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_WindowWithCount_ObservableError
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					observe := testObservable[int](ctx, 1, 2, errFoo, 4, 5).WindowWithCount(2).Observe()

					rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
						rx.HasItems[int]{
							Expected: []int{1, 2},
						},
					)

					rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
						rx.IsEmpty[int]{},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})

			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_WindowWithCount_InputError
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := rx.Empty[int]().WindowWithCount(-1)
					rx.Assert(ctx, obs,
						rx.HasAnError[int]{},
					)
				})
			})
		})
	})

	Context("WindowWithTime", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_WindowWithTime
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int], 10)
				ch <- rx.Of(1)
				ch <- rx.Of(2)
				obs := rx.FromChannel(ch)

				go func() {
					time.Sleep(30 * time.Millisecond)
					ch <- rx.Of(3)
					close(ch)
				}()

				observe := obs.WindowWithTime(
					rx.WithDuration(10*time.Millisecond),
					rx.WithBufferedChannel[int](10),
				).Observe()

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItems[int]{
						Expected: []int{1, 2},
					},
				)

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItem[int]{
						Expected: 3,
					},
				)
			})
		})
	})

	Context("WindowWithTimeOrCount", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_WindowWithTimeOrCount
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int], 10)
				ch <- rx.Of(1)
				ch <- rx.Of(2)
				obs := rx.FromChannel(ch)

				go func() {
					time.Sleep(30 * time.Millisecond)
					ch <- rx.Of(3)
					close(ch)
				}()

				observe := obs.WindowWithTimeOrCount(
					rx.WithDuration(10*time.Millisecond), 1,
					rx.WithBufferedChannel[int](10),
				).Observe()

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItem[int]{
						Expected: 1,
					},
				)

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItem[int]{
						Expected: 2,
					},
				)

				rx.Assert(ctx, (<-observe).Opaque().(rx.Observable[int]),
					rx.HasItem[int]{
						Expected: 3,
					},
				)
			})
		})
	})
})
