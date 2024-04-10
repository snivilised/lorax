package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
)

var _ = Describe("Observable operator", func() {
	Context("DoOnComplete", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnCompleted_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				called := false
				<-testObservable[int](ctx, 1, 2, 3).DoOnCompleted(func() {
					called = true
				})

				Expect(called).To(BeTrue())
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnCompleted_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					called := false
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnCompleted(func() {
						called = true
					})

					Expect(called).To(BeTrue())
				})
			})
		})
	})

	Context("DoOnError", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnError_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var got error
				<-testObservable[int](ctx, 1, 2, 3).DoOnError(func(err error) {
					got = err
				})

				Expect(got).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnError_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					var got error
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnError(func(err error) {
						got = err
					})

					Expect(got).To(Equal(errFoo))
				})
			})
		})
	})

	Context("DoOnNext", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnNext_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				s := make([]int, 0)
				<-testObservable[int](ctx, 1, 2, 3).DoOnNext(func(i int) {
					s = append(s, i)
				})

				Expect(s).To(ContainElements([]int{1, 2, 3}))
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnNext_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					s := make([]int, 0)
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnNext(func(i int) {
						s = append(s, i)
					})

					Expect(s).To(ContainElements([]int{1}))
				})
			})
		})
	})
})
