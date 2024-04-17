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

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Distinct", func() {
		When("duplicates present", func() {
			It("🧪 should: suppress duplicates", func() {
				// rxgo: Test_Observable_Distinct
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 2, 1, 3).Distinct(
					func(_ context.Context, value int) (int, error) {
						return value, nil
					},
				)

				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{1, 2, 3},
					},
					rx.HasNoError[int]{},
				)
			})
		})

		Context("Errors", func() {
			When("error present", func() {
				It("🧪 should: emit values before error and has error", func() {
					// rxgo: Test_Observable_Distinct_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, errFoo, 3).Distinct(
						func(_ context.Context, value int) (int, error) {
							return value, nil
						},
					)

					rx.Assert(ctx, obs,
						rx.ContainItems[int]{
							Expected: []int{1, 2},
						},
						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})

			When("error present", func() {
				It("🧪 should: emit values before error and has error", func() {
					// rxgo: Test_Observable_Distinct_Error2
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, 2, 3, 4).Distinct(
						func(_ context.Context, value int) (int, error) {
							if value == 3 {
								return 0, errFoo
							}

							return value, nil
						},
					)

					rx.Assert(ctx, obs,
						rx.ContainItems[int]{
							Expected: []int{1, 2},
						}, rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})

		Context("Parallel", func() {
			Context("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Distinct_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, 1, 3).Distinct(
						func(_ context.Context, item int) (int, error) {
							return item, nil
						}, rx.WithCPUPool[int]())

					rx.Assert(ctx, obs,
						rx.HasItemsNoOrder[int]{
							Expected: []int{1, 2, 3},
						},
						rx.HasNoError[int]{},
					)
				})
			})
		})

		Context("Parallel/Error", func() {
			When("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Distinct_Parallel_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, errFoo).Distinct(
						func(_ context.Context, item int) (int, error) {
							return item, nil
						}, rx.WithContext[int](ctx), rx.WithCPUPool[int]())

					rx.Assert(ctx, obs, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})

			When("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Distinct_Parallel_Error2
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, 2, 3, 4).Distinct(
						func(_ context.Context, item int) (int, error) {
							if item == 3 {
								return 0, errFoo
							}
							return item, nil
						}, rx.WithContext[int](ctx), rx.WithCPUPool[int](),
					)

					rx.Assert[int](ctx, obs, rx.HasError[int]{
						Expected: []error{errFoo},
					})
				})
			})
		})
	})

	Context("DistinctUntilChanged", func() {
		Context("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DistinctUntilChanged
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 2, 1, 3).DistinctUntilChanged(
					func(_ context.Context, v int) (int, error) {
						return v, nil
					}, rx.NativeItemLimitComparator, rx.WithCPUPool[int]())

				rx.Assert(ctx, obs,
					rx.ContainItems[int]{
						Expected: []int{1, 2, 1, 3},
					})
			})
		})

		Context("Parallel", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DistinctUntilChanged_Parallel
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, 2, 2, 1, 3).DistinctUntilChanged(
						func(_ context.Context, value int) (int, error) {
							return value, nil
						}, rx.NativeItemLimitComparator, rx.WithCPUPool[int]())

					rx.Assert(ctx, obs, rx.ContainItems[int]{
						Expected: []int{1, 2, 1, 3},
					})
				})
			})
		})
	})
})
