package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("ForEach", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_ForEach_Done
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var gotErr error
				count := 0
				done := make(chan struct{})
				obs := testObservable[int](ctx, 1, 2, 3)
				obs.ForEach(func(i rx.Item[int]) {
					count += i.V
				}, func(err error) {
					gotErr = err
					done <- struct{}{}
				}, func() {
					done <- struct{}{}
				})

				// We avoid using the assertion API on purpose
				<-done

				Expect(count).To(Equal(6))
				Expect(gotErr).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_ForEach_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					count := 0
					var gotErr error
					done := make(chan struct{})

					obs := testObservable[int](ctx, 1, 2, 3, errFoo)
					obs.ForEach(func(i rx.Item[int]) {
						count += i.V
					}, func(err error) {
						gotErr = err
						select {
						case <-ctx.Done():
							return
						case done <- struct{}{}:
						}
					}, func() {
						select {
						case <-ctx.Done():
							return
						case done <- struct{}{}:
						}
					}, rx.WithContext[int](ctx))

					// We avoid using the assertion API on purpose
					<-done

					Expect(count).To(Equal(6))
					Expect(gotErr).To(MatchError(errFoo))
				})
			})
		})
	})
})
