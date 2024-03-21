package boost_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok

	"github.com/snivilised/lorax/boost"
)

var _ = Describe("AnnotatedWaitGroup", func() {

	Context("Add", func() {
		It("should: add", func() {
			wg := boost.NewAnnotatedWaitGroup("add-unit-test")

			wg.Add(1, "producer")
		})
	})

	Context("Done", func() {
		It("should: quit", func() {
			wg := boost.NewAnnotatedWaitGroup("done-unit-test")

			wg.Add(1, "producer")
			wg.Done("producer")
		})
	})

	Context("Wait", func() {
		It("should: quit", func() {
			wg := boost.NewAnnotatedWaitGroup("wait-unit-test")

			wg.Add(1, "producer")
			go func() {
				wg.Done("producer")
			}()
			<-time.After(time.Second / 10)
			wg.Wait("main")
		})
	})
})
