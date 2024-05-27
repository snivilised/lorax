package boost_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/boost"
)

var _ = Describe("WorkerPoolFunc", func() {
	Context("ants", func() {
		It("should: not fail", func() {
			// TestNonblockingSubmit
			var wg sync.WaitGroup

			pool, err := boost.NewFuncPool[int, int](AntsSize, demoPoolFunc, &wg)

			defer pool.Release()

			for i := 0; i < n; i++ { // producer
				_ = pool.Post(Param)
			}
			wg.Wait()
			GinkgoWriter.Printf("pool with func, running workers number:%d\n",
				pool.Running(),
			)
			ShowMemStats()

			Expect(err).To(Succeed())
		})
	})
})
