package rx_test

import (
	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	XContext("${{OPERATOR-NAME}}", func() {
		Context("principle", func() {
			// success path
			It("🧪 should: ", func() {
				// rxgo: Test_
				defer leaktest.Check(GinkgoT())()

				Expect(1).To(Equal(1))
				rx.Just("delete-me")
			})
		})

		Context("Errors", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_
					defer leaktest.Check(GinkgoT())()
				})
			})
		})

		Context("Parallel", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_
					defer leaktest.Check(GinkgoT())()
				})
			})
		})

		Context("Parallel/Error", func() {
			Context("given: foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_
					defer leaktest.Check(GinkgoT())()
				})
			})
		})
	})
})
