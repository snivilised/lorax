package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
	"github.com/snivilised/lorax/rxa"
)

var _ = Describe("Item", Ordered, func() {
	Context("SendItems", func() {
		Context("variadic", func() {
			When("no errors in observable", func() {
				It("🧪 should: send items without error", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)

					rx.SendItems(context.Background(), ch, rx.CloseChannel,
						rx.Of(1),
						rx.Of(2),
						rx.Of(3),
					)

					rxa.Assert(context.Background(),
						rx.FromChannel(ch),
						rxa.HasItems([]int{1, 2, 3}),
						rxa.HasNoError[int]())
				})
			})

			When("error in observable", func() {
				It("🧪 should: send items including error", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)

					rx.SendItems(context.Background(), ch, rx.CloseChannel,
						rx.Of(1),
						rx.Error[int](errFoo),
						rx.Of(3),
					)

					rxa.Assert(context.Background(),
						rx.FromChannel(ch),
						rxa.HasItems([]int{1, 3}),
						rxa.HasAnError[int]())
				})
			})

			When("specific error in observable", func() {
				It("🧪 should: send items including error", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)

					rx.SendItems(context.Background(), ch, rx.CloseChannel,
						rx.Of(1),
						rx.Error[int](errFoo),
						rx.Of(3),
					)

					rxa.Assert(context.Background(),
						rx.FromChannel(ch),
						rxa.HasItems([]int{1, 3}),
						rxa.HasError[int](errFoo))
				})
			})
		})

		Context("blocking", func() {
			When("no errors in observable", func() {
				It("foo", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)
					rx.Of[int](5).SendBlocking(ch)

					Expect((<-ch).V).To(Equal(5))
				})
			})
		})

		Context("context", func() {
			When("not cancelled", func() {
				It("🧪 should: return true", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					Expect(rx.Of(5).SendContext(ctx, ch)).To(BeTrue())
				})
			})

			When("cancelled", func() {
				It("🧪 should: return false", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)

					ctx, cancel := context.WithCancel(context.Background())
					cancel()

					Expect(rx.Of(5).SendContext(ctx, ch)).To(BeFalse())
				})
			})
		})

		Context("non-blocking", func() {
			When("channel free", func() {
				It("🧪 should: send item and return true", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)

					Expect(rx.Of(5).SendNonBlocking(ch)).To(BeTrue())
				})
			})

			When("channel busy", func() {
				It("🧪 should: not send item and return false", func() {
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)

					rx.Of(5).SendNonBlocking(ch)
					Expect(rx.Of(5).SendNonBlocking(ch)).To(BeFalse())
				})
			})
		})
	})
})
