package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/enums"
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Item", Ordered, func() {
	Context("SendItems", func() {
		Context("variadic", func() {
			When("no errors in observable", func() {
				It("🧪 should: send items without error", func() {
					// Test_SendItems_Variadic
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)

					rx.SendItems(context.Background(), ch, enums.CloseChannel,
						1, 2, 3,
					)

					rx.Assert(context.Background(),
						rx.FromChannel(ch),
						rx.HasItems[int]{
							Expected: []int{1, 2, 3},
						},
						rx.HasNoError[int]{},
					)
				})
			})

			When("error in observable", func() {
				It("🧪 should: send items including error", func() {
					// Test_SendItems_VariadicWithError
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)
					rx.SendItems(context.Background(), ch, enums.CloseChannel,
						1,
						rx.Error[int](errFoo),
						3,
					)

					rx.Assert(context.Background(),
						rx.FromChannel(ch),
						rx.HasItems[int]{
							Expected: []int{1, 3},
						},
						rx.HasAnError[int]{},
					)
				})
			})

			When("slice", func() {
				It("🧪 should: send slice", func() {
					// Test_SendItems_Slice
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)
					go rx.SendItems(context.Background(), ch, enums.CloseChannel, []int{1, 2, 3})
					rx.Assert(context.Background(), rx.FromChannel(ch),
						rx.HasItems[int]{
							Expected: []int{1, 2, 3},
						},

						rx.HasNoError[int]{},
					)
				})
			})

			When("specific error observed", func() {
				It("🧪 should: send items including error", func() {
					// Test_SendItems_SliceWithError
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 3)
					go rx.SendItems(context.Background(), ch, enums.CloseChannel, []any{1, errFoo, 3})
					rx.Assert(context.Background(), rx.FromChannel(ch),
						rx.HasItems[int]{
							Expected: []int{1, 3},
						},

						rx.HasError[int]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})

		Context("blocking", func() {
			When("no errors in observable", func() {
				It("foo", func() {
					// Test_Item_SendBlocking
					defer leaktest.Check(GinkgoT())()

					ch := make(chan rx.Item[int], 1)
					defer close(ch)

					rx.Of(5).SendBlocking(ch)
					Expect((<-ch).V).To(Equal(5))
				})
			})
		})

		Context("context", func() {
			When("not cancelled", func() {
				It("🧪 should: return true", func() {
					// Test_Item_SendContext_True
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
					// Test_Item_SendContext_False
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
				// Test_Item_SendNonBlocking
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
