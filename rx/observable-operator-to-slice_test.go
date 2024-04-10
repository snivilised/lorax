package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("ToSlice", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_ToSlice
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				s, err := testObservable[int](ctx, 1, 2, 3).ToSlice(5)

				Expect(s).To(ContainElements([]rx.Item[int]{
					rx.Of(1),
					rx.Of(2),
					rx.Of(3),
				}))

				Expect(s).To(HaveCap(5))
				Expect(err).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_ToSlice_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					s, err := testObservable[int](ctx, 1, 2, errFoo, 3).ToSlice(0)

					Expect(s).To(ContainElements([]rx.Item[int]{
						rx.Of(1),
						rx.Of(2),
					}))

					Expect(err).To(MatchError(errFoo))
				})
			})
		})
	})
})
