package boost_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok

	"github.com/snivilised/lorax/boost"
	"github.com/snivilised/lorax/internal/ants"
)

var _ = Describe("WorkerPoolSubmitter", func() {
	Context("ants", func() {
		It("should: not fail", func() {
			// TestNonblockingSubmit
			var wg sync.WaitGroup

			pool, err := boost.NewSubmitterPool[int, int](PoolSize, &wg,
				ants.WithNonblocking(true),
			)
			defer pool.Release()

			Expect(err).To(Succeed())
			Expect(pool).NotTo(BeNil())

			for i := 0; i < PoolSize-1; i++ {
				Expect(pool.Post(longRunningFunc)).To(Succeed(),
					"nonblocking submit when pool is not full shouldn't return error",
				)
			}

			firstCh := make(chan struct{})
			secondCh := make(chan struct{})
			fn := func() {
				<-firstCh
				close(secondCh)
			}
			// pool is full now.
			Expect(pool.Post(fn)).To(Succeed(),
				"nonblocking submit when pool is not full shouldn't return error",
			)
			Expect(pool.Post(demoFunc)).To(MatchError(ants.ErrPoolOverload.Error()),
				"nonblocking submit when pool is full should get an ErrPoolOverload",
			)
			// interrupt fn to get an available worker
			close(firstCh)
			<-secondCh
			Expect(pool.Post(demoFunc)).To(Succeed(),
				"nonblocking submit when pool is not full shouldn't return error",
			)
		})
	})
})
