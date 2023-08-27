package async_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/snivilised/lorax/async"
)

var _ = Describe("AnnotatedWaitGroup", func() {

	Context("Add", func() {
		It("should: add", func() {
			wg := async.NewAnnotatedWaitGroup("add-unit-test")

			wg.Add(1, "producer")
		})
	})

	Context("Done", func() {
		It("should: quit", func() {
			wg := async.NewAnnotatedWaitGroup("done-unit-test")

			wg.Add(1, "producer")
			wg.Done("producer")
		})
	})

	Context("Wait", func() {
		It("should: quit", func() {
			wg := async.NewAnnotatedWaitGroup("wait-unit-test")

			wg.Add(1, "producer")
			go func() {
				wg.Done("producer")
			}()
			<-time.After(time.Second / 10)
			wg.Wait("main")
		})
	})
})
