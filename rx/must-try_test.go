package rx_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/rx"
)

func expectTryError() {
	pe := recover()
	if err, ok := pe.(error); !ok || !strings.Contains(err.Error(),
		"expected item to be") {
		Fail("incorrect error")
	}
}

var _ = Describe("try", func() {
	Context("success", func() {
		When("Bool", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Bool[float32](true)
				result, err := rx.TryBool(item)

				Expect(err).To(Succeed())
				Expect(result).To(BeTrue())
			})
		})

		When("WCh", func() {
			It("🧪 should: get value ok", func() {
				item := rx.WCh[float32](make(chan<- rx.Item[float32]))
				result, err := rx.TryWCh(item)

				Expect(err).To(Succeed())
				Expect(result).NotTo(BeNil())
			})
		})

		When("Num", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Num[float32](42)
				result, err := rx.TryNum(item)

				Expect(err).To(Succeed())
				Expect(result).To(Equal(42))
			})
		})

		When("Opaque", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Opaque[float32](widget{
					name:   "kramer",
					amount: 10,
				})
				result, err := rx.TryOpaque[float32, widget](item)

				Expect(err).To(Succeed())
				Expect(result).To(Equal(widget{
					name:   "kramer",
					amount: 10,
				}))
			})
		})

		When("TV", func() {
			It("🧪 should: get value ok", func() {
				item := rx.TV[float32](42)

				Expect(rx.MustTV(item)).To(Equal(42))
			})
		})
	})

	Context("fail", func() {
		When("Bool", func() {
			It("🧪 should: result in error", func() {
				item := rx.Of[float32](42)
				_, err := rx.TryBool(item)

				Expect(err).NotTo(Succeed())
			})
		})

		When("WCh", func() {
			It("🧪 should: result in error", func() {
				item := rx.Of[float32](42)
				_, err := rx.TryWCh(item)

				Expect(err).NotTo(Succeed())
			})
		})

		When("Num", func() {
			It("🧪 should: result in error", func() {
				item := rx.Of[float32](42)
				_, err := rx.TryNum(item)

				Expect(err).NotTo(Succeed())
			})
		})

		When("Opaque", func() {
			It("🧪 should: result in error", func() {
				item := rx.Of[float32](42)
				_, err := rx.TryOpaque[float32, widget](item)

				Expect(err).NotTo(Succeed())
			})
		})

		When("TV", func() {
			It("🧪 should: result in error", func() {
				item := rx.Of[float32](42)
				_, err := rx.TryTV(item)

				Expect(err).NotTo(Succeed())
			})
		})
	})
})

var _ = Describe("must", func() {
	Context("success", func() {
		When("Bool", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Bool[float32](true)

				Expect(rx.MustBool(item)).To(BeTrue())
			})
		})

		When("WCh", func() {
			It("🧪 should: get value ok", func() {
				item := rx.WCh[float32](make(chan<- rx.Item[float32]))

				Expect(rx.MustWCh(item)).NotTo(BeNil())
			})
		})

		When("Num", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Num[float32](42)

				Expect(rx.MustNum(item)).To(Equal(42))
			})
		})

		When("Opaque", func() {
			It("🧪 should: get value ok", func() {
				item := rx.Opaque[float32](widget{
					name:   "kramer",
					amount: 10,
				})

				Expect(rx.MustOpaque[float32, widget](item)).To(Equal(widget{
					name:   "kramer",
					amount: 10,
				}))
			})
		})

		When("TV", func() {
			It("🧪 should: get value ok", func() {
				item := rx.TV[float32](42)

				Expect(rx.MustTV(item)).To(Equal(42))
			})
		})
	})

	Context("fail", func() {
		When("Bool", func() {
			It("🧪 should: panic", func() {
				defer expectTryError()

				item := rx.Of[float32](42)
				rx.MustBool(item)

				Fail("❌ expected panic due to incorrect must try attempt")
			})
		})

		When("WCh", func() {
			It("🧪 should: panic", func() {
				defer expectTryError()

				item := rx.Of[float32](42)
				rx.MustWCh(item)

				Fail("❌ expected panic due to incorrect must try attempt")
			})
		})

		When("Num", func() {
			It("🧪 should: panic", func() {
				defer expectTryError()

				item := rx.Of[float32](42)
				rx.MustNum(item)

				Fail("❌ expected panic due to incorrect must try attempt")
			})
		})

		When("Opaque", func() {
			It("🧪 should: panic", func() {
				defer expectTryError()

				item := rx.Of[float32](42)
				rx.MustOpaque[float32, widget](item)

				Fail("❌ expected panic due to incorrect must try attempt")
			})
		})

		When("TV", func() {
			It("🧪 should: panic", func() {
				defer expectTryError()

				item := rx.Of[float32](42)
				rx.MustTV(item)

				Fail("❌ expected panic due to incorrect must try attempt")
			})
		})
	})
})
