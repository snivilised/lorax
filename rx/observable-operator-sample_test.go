package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Sample", func() {
		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Sample_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1).Sample(rx.Empty[int](), rx.WithContext[int](ctx))
				rx.Assert(ctx, obs,
					rx.IsEmpty[int]{},
					rx.HasNoError[int]{},
				)
			})
		})
	})
})
