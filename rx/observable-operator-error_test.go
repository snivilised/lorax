package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/enums"
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Error", func() {
		When("no error", func() {
			It("🧪 should: return nil", func() {
				// rxgo: Test_Observable_Error_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3)
				Expect(obs.Error()).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: return encountered error", func() {
					// rxgo: Test_Observable_Error_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[int](ctx, 1, errFoo, 3)
					Expect(obs.Error()).Error().To(MatchError(errFoo))
				})
			})
		})
	})

	Context("Errors", func() {
		When("one error", func() {
			It("🧪 should: contain just 1 error", func() {
				// rxgo: Test_Observable_Errors_OneError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, errFoo, 3)
				Expect(obs.Errors()).To(HaveLen(1))
			})
		})

		When("multiple errors", func() {
			It("🧪 should: return all encountered errors", func() {
				// rxgo: Test_Observable_Errors_MultipleError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, errFoo, errBar)
				Expect(obs.Errors()).To(HaveLen(2))
			})
		})

		When("multiple errors from map", func() {
			It("🧪 should: return all encountered errors", func() {
				// rxgo: Test_Observable_Errors_MultipleErrorFromMap
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 2, 3, 4).Map(
					func(_ context.Context, i int) (int, error) {
						if i == 2 {
							return 0, errFoo
						}
						if i == 3 {
							return 0, errBar
						}
						return i, nil
					}, rx.WithErrorStrategy[int](enums.ContinueOnError),
				)
				Expect(obs.Errors()).To(HaveLen(2))
			})
		})
	})
})
