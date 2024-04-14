package rx_test

import (
	"context"
	"errors"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Repeat", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Repeat
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				repeat := testObservable[int](ctx, 1, 2, 3).Repeat(1, nil)
				rx.Assert(ctx, repeat, rx.HasItems[int]{
					Expected: []int{1, 2, 3, 1, 2, 3},
				})
			})
		})

		When("zero", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Repeat_Zero
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				repeat := testObservable[int](ctx, 1, 2, 3).Repeat(0, nil)
				rx.Assert(ctx, repeat, rx.HasItems[int]{
					Expected: []int{1, 2, 3},
				})
			})
		})

		When("infinite", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Repeat_Infinite
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				repeat := testObservable[int](ctx, 1, 2, 3).Repeat(
					rx.Infinite, nil, rx.WithContext[int](ctx),
				)

				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()

				rx.Assert(ctx, repeat, rx.HasNoError[int]{},
					rx.CustomPredicate[int]{
						Expected: func(actual rx.AssertResources[int]) error {
							items := actual.Values()
							if len(items) == 0 {
								return errors.New("no items")
							}

							return nil
						},
					})
			})
		})

		When("frequency", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Repeat_Frequency
				defer leaktest.Check(GinkgoT())()

				/*
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					frequency := new(mockDuration)
					frequency.On("duration").Return(time.Millisecond)

					repeat := testObservable[int](ctx, 1, 2, 3).Repeat(1, frequency)
					rx.Assert(ctx, repeat, rx.HasItems[int]{
						Expected: []int{1, 2, 3, 1, 2, 3},
					})

					frequency.AssertNumberOfCalls("duration", 1)
					frequency.AssertExpectations()
				*/
			})
		})

		Context("Errors", func() {
			When("Negative Count", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Repeat_NegativeCount
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					repeat := testObservable[int](ctx, 1, 2, 3).Repeat(-2, nil)
					rx.Assert(ctx, repeat,
						rx.IsEmpty[int]{},
						rx.HasAnError[int]{},
					)
				})
			})
		})
	})
})