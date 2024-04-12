package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
)

var _ = Describe("Observable operator", func() {
	Context("Run", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Run
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				s := make([]int, 0)
				<-testObservable[int](ctx, 1, 2, 3).Map(
					func(_ context.Context, i int) (int, error) {
						s = append(s, i)

						return i, nil
					},
				).Run()

				Expect(s).To(ContainElements([]int{1, 2, 3}))
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Run_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					s := make([]int, 0)
					<-testObservable[int](ctx, 1, errFoo).Map(
						func(_ context.Context, i int) (int, error) {
							s = append(s, i)

							return i, nil
						},
					).Run()

					Expect(s).To(ContainElements([]int{1}))
				})
			})
		})
	})
})
