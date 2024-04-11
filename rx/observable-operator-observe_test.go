package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
)

var _ = Describe("Observable operator", func() {
	Context("Observe", func() {
		When("observable", func() {
			It("🧪 should: receive all emitted items", func() {
				// rxgo: Test_Observable_Observe
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				got := make([]int, 0)
				ch := testObservable[int](ctx, 1, 2, 3).Observe()
				for item := range ch {
					got = append(got, item.V)
				}
				Expect(got).To(ContainElements([]int{1, 2, 3}))
			})
		})
	})
})
