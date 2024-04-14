package rx_test

import (
	"context"
	"fmt"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("TimeInterval", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_TimeInterval
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3).TimeInterval()

				rx.Assert(ctx, obs, rx.CustomPredicate[int]{
					Expected: func(actual rx.AssertResources[int]) error {
						items := actual.Opaques()

						if len(items) != 3 {
							return fmt.Errorf("expected 3 items, got %d items", len(items))
						}

						return nil
					},
				})
			})
		})
	})

	Context("Timestamp", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Timestamp
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				observe := testObservable[int](ctx, 1, 2, 3).Timestamp().Observe()
				v, _ := (<-observe).O.(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(1))

				v, _ = (<-observe).O.(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(2))

				v, _ = (<-observe).O.(*rx.TimestampItem[int])
				Expect(v.V).To(Equal(3))
			})
		})
	})
})
