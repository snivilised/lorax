package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Sum", func() {
		When("principle", func() {
			It("🧪 should: return sum", func() {
				// rxgo: Test_Observable_SumFloat32_OnlyFloat32
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx, testObservable[float32](ctx,
					// explicit cast to float32 is required to match
					// the generic type to ensure observable sends the
					// item. Failure to perform this task results in
					// an error ("channel value: '1' not sent (wrong type?)").
					//
					float32(1.0), float32(2.0), float32(3.0),
				).Sum(rx.Calc[float32]()),
					rx.HasItem[float32]{
						Expected: 6.0,
					})
			})
		})

		When("empty", func() {
			It("🧪 should: return empty result", func() {
				// rxgo: Test_Observable_SumFloat32_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				rx.Assert(ctx,
					rx.Empty[float32]().Sum(rx.Calc[float32]()),
					rx.IsEmpty[float32]{},
				)
			})
		})

		// NB(Test_Observable_SumFloat32_DifferentTypes): Sum does
		// not support different types, only values of type T are
		// supported.

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: result in error", func() {
					// rxgo: Test_Observable_SumFloat32_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					rx.Assert(ctx, testObservable[float32](ctx,
						// we omit the explicit cast to float32 as normally would
						// be the case, to enforce a resulting error.
						//
						1.1, 2.2, 3.3,
					).Sum(rx.Calc[float32]()),
						rx.HasAnError[float32]{},
					)
				})
			})
		})
	})
})
