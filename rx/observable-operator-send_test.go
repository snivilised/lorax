package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("Send", func() {
		When("channel is buffered", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Send
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int], 10)
				testObservable[int](ctx, 1, 2, 3, errFoo).Send(ch)
				Expect(rx.Of(1)).To(Equal(<-ch))
				Expect(rx.Of(2)).To(Equal(<-ch))
				Expect(rx.Of(3)).To(Equal(<-ch))
				Expect(rx.Error[int](errFoo)).To(Equal(<-ch))
			})
		})

		When("channel is not buffered", func() {
			It("🧪 should: ", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int])
				testObservable[int](ctx, 1, 2, 3, errFoo).Send(ch)
				Expect(rx.Of(1)).To(Equal(<-ch))
				Expect(rx.Of(2)).To(Equal(<-ch))
				Expect(rx.Of(3)).To(Equal(<-ch))
				Expect(rx.Error[int](errFoo)).To(Equal(<-ch))
			})
		})
	})
})
