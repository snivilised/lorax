package rx_test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("GroupBy", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_GroupBy
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				length := 3
				count := 11
				obs := rx.Range[int](0, count).GroupBy(length, func(item rx.Item[int]) int {
					return item.Num() % length
				}, rx.WithBufferedChannel[int](count))
				observables, err := obs.ToSlice(0)

				if err != nil {
					Fail(err.Error())
				}

				if len(observables) != length {
					Fail(fmt.Sprintf("length; got=%d, expected=%d", len(observables), length))
				}

				rx.Assert(ctx, observables[0].Opaque().(rx.Observable[int]),
					rx.HasNumbers[int]{
						Expected: []int{0, 3, 6, 9},
					},
					rx.HasNoError[int]{},
				)
				rx.Assert(ctx, observables[1].Opaque().(rx.Observable[int]),
					rx.HasNumbers[int]{
						Expected: []int{1, 4, 7, 10},
					},
					rx.HasNoError[int]{},
				)
				rx.Assert(ctx, observables[2].Opaque().(rx.Observable[int]),
					rx.HasNumbers[int]{
						Expected: []int{2, 5, 8},
					},
					rx.HasNoError[int]{},
				)
			})
		})

		When("dynamic", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_GroupByDynamic
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				length := 3
				count := 11

				obs := rx.Range[int](0, count).GroupByDynamic(func(item rx.Item[int]) string {
					if item.Num() == 10 {
						return "10"
					}

					return strconv.Itoa(item.Num() % length)
				}, rx.WithBufferedChannel[int](count))
				observablesGrouped, err := obs.ToSlice(0)

				if err != nil {
					Fail(err.Error())
				}

				if len(observablesGrouped) != 4 {
					Fail(fmt.Sprintf("length; got=%d, expected=%d", len(observablesGrouped), 4))
				}

				rx.Assert(ctx, observablesGrouped[0].Opaque().(rx.GroupedObservable[int]),
					rx.HasNumbers[int]{
						Expected: []int{0, 3, 6, 9},
					},
					rx.HasNoError[int]{},
				)
				Expect(observablesGrouped[0].Opaque().(rx.GroupedObservable[int]).Key).To(Equal("0"))

				rx.Assert(ctx, observablesGrouped[1].Opaque().(rx.GroupedObservable[int]),
					rx.HasNumbers[int]{
						Expected: []int{1, 4, 7},
					},
					rx.HasNoError[int]{},
				)
				Expect(observablesGrouped[1].Opaque().(rx.GroupedObservable[int]).Key).To(Equal("1"))

				rx.Assert(ctx, observablesGrouped[2].Opaque().(rx.GroupedObservable[int]),
					rx.HasNumbers[int]{
						Expected: []int{2, 5, 8},
					},
					rx.HasNoError[int]{},
				)
				Expect(observablesGrouped[2].Opaque().(rx.GroupedObservable[int]).Key).To(Equal("2"))

				rx.Assert(ctx, observablesGrouped[3].Opaque().(rx.GroupedObservable[int]),
					rx.HasNumbers[int]{
						Expected: []int{10},
					},
					rx.HasNoError[int]{},
				)
				Expect(observablesGrouped[3].Opaque().(rx.GroupedObservable[int]).Key).To(Equal("10"))
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_GroupBy_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					length := 3
					count := 11

					obs := rx.Range[int](0, count).GroupBy(length, func(_ rx.Item[int]) int {
						return 4
					}, rx.WithBufferedChannel[int](count))
					observables, err := obs.ToSlice(0)

					if err != nil {
						Fail(err.Error())
					}

					if len(observables) != length {
						Fail(fmt.Sprintf("length; got=%d, expected=%d", len(observables), length))
					}

					rx.Assert(ctx, observables[0].Opaque().(rx.Observable[int]),
						rx.HasAnError[int]{},
					)
					rx.Assert(ctx, observables[1].Opaque().(rx.Observable[int]),
						rx.HasAnError[int]{},
					)
					rx.Assert(ctx, observables[2].Opaque().(rx.Observable[int]),
						rx.HasAnError[int]{},
					)
				})
			})
		})
	})
})
