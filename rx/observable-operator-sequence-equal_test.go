package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("SequenceEqual", func() {
		When("even sequence", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_EvenSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequence := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213)
				result := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequence,
					rx.NumericItemLimitComparator,
				)
				rx.Assert(ctx, result,
					rx.IsTrue[int]{},
				)
			})
		})

		When("Uneven sequence", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_UnevenSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequence := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213)
				result := testObservable[int](ctx, 2, 5, 12, 43, 15, 100, 213).SequenceEqual(
					sequence,
					rx.NumericItemLimitComparator,
					rx.WithContext[int](ctx),
				)
				rx.Assert(ctx, result,
					rx.IsTrue[int]{},
				)
			})
		})

		When("Different sequence length", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_DifferentLengthSequence
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				sequenceShorter := testObservable[int](ctx, 2, 5, 12, 43, 98, 100)
				sequenceLonger := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213, 512)

				resultForShorter := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequenceShorter,
					rx.NumericItemLimitComparator,
				)
				rx.Assert(ctx, resultForShorter,
					rx.IsFalse[int]{},
				)

				resultForLonger := testObservable[int](ctx, 2, 5, 12, 43, 98, 100, 213).SequenceEqual(
					sequenceLonger,
					rx.NumericItemLimitComparator,
				)
				rx.Assert(ctx, resultForLonger,
					rx.IsFalse[int]{},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_SequenceEqual_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				result := rx.Empty[int]().SequenceEqual(
					rx.Empty[int](),
					rx.NumericItemLimitComparator,
				)
				rx.Assert(ctx, result,
					rx.IsTrue[int]{},
				)
			})
		})
	})
})
