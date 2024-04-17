package rx_test

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
				rx.Assert(ctx, repeat, rx.ContainItems[int]{
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
				rx.Assert(ctx, repeat, rx.ContainItems[int]{
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
