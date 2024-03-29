package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/onsi/ginkgo/v2/dsl/decorators"
	. "github.com/onsi/gomega" //nolint:revive // gomega ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("OptionSingle", func() {
	Context("Get", func() {
		When("Just", func() {
			It("🧪 should: get the single value", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.NewOptionalSingleImpl(rx.Just(1)())
				get, err := single.Get()
				Expect(err).Error().To(BeNil())
				Expect(get.V).To(Equal(1))
			})
		})

		When("JustItem", func() {
			It("🧪 should: get the single value", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.NewOptionalSingleImpl(rx.JustItem(1))
				get, err := single.Get()
				Expect(err).Error().To(BeNil())
				Expect(get.V).To(Equal(1))
			})
		})

		When("JustSingle", func() {
			It("🧪 should: get the single value", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.NewOptionalSingleImpl(rx.JustSingle(1)())
				get, err := single.Get()
				Expect(err).Error().To(BeNil())
				Expect(get.V).To(Equal(1))
			})
		})

		When("Empty", func() {
			It("🧪 should: get empty item", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.NewOptionalSingleImpl(rx.Empty[int]())
				get, err := single.Get()
				Expect(err).Error().To(BeNil())
				Expect(get).To(Equal(rx.Item[int]{}))
			})
		})

		When("Error", func() {
			It("🧪 should: get error value", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.NewOptionalSingleImpl(rx.JustError[int](errFoo)())
				get, err := single.Get()
				Expect(err).Error().To(BeNil())
				Expect(get.E).To(Equal(errFoo))
			})
		})

		When("Context Cancelled", func() {
			It("🧪 should: result in cancellation error", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				single := rx.NewOptionalSingleImpl(rx.Never[int]())
				cancel()
				_, err := single.Get(rx.WithContext[int](ctx))
				Expect(ctx.Err()).Error().To(Equal(err))
			})
		})
	})

	Context("Map", Ordered, func() {
		var increment rx.Func[int]

		BeforeAll(func() {
			increment = func(_ context.Context, i int) (int, error) {
				return i + 1, nil
			}
		})

		When("Just", func() {
			It("🧪 should: Map the single entity iterator", func() {
				defer leaktest.Check(GinkgoT())()

				single := rx.Just(42)().Max(rx.LimitComparator).Map(increment)
				rx.Assert(context.Background(), single, rx.HasItem(43), rx.HasNoError[int]())
			})

			Context("Max", decorators.Label("comprehension"), func() {
				It("🧪 should: turn the sequence into a Single iterable", func() {
					defer leaktest.Check(GinkgoT())()

					single := rx.Just(42, 48)().Max(rx.LimitComparator)
					rx.Assert(context.Background(), single, rx.HasItem(48), rx.HasNoError[int]())
				})
			})

			Context("Min", decorators.Label("comprehension"), func() {
				It("🧪 should: turn the sequence into a Single iterable", func() {
					defer leaktest.Check(GinkgoT())()

					single := rx.Just(42, 48)().Min(rx.LimitComparator)
					rx.Assert(context.Background(), single, rx.HasItem(42), rx.HasNoError[int]())
				})
			})
		})
	})

	Context("Observe", func() {
		When("JustItem", func() {
			It("🧪 should: who knows", func() {
				defer leaktest.Check(GinkgoT())()
				// the intentions of the original rxgo test is not particularly clear
				//
				single := rx.JustItem(42).Filter(func(i int) bool {
					return i == 42
				})
				rx.Assert(context.Background(), single, rx.HasItem(42), rx.HasNoError[int]())
			})

			Context("Filter", decorators.Label("comprehension"), func() {
				When("item filtered out", func() {
					It("should: result in empty single iterable", func() {
						defer leaktest.Check(GinkgoT())()

						single := rx.JustItem(42).Filter(func(i int) bool {
							return i == 48
						})
						rx.Assert(context.Background(), single, rx.IsEmpty[int](), rx.HasNoError[int]())
					})
				})
			})
		})
	})
})
