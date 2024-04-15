package rx_test

import (
	"context"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("ZipFromIterable", func() {
		When("source and other observers same length", func() {
			It("🧪 should: emit zipped elements", func() {
				// rxgo: Test_Observable_ZipFromObservable
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2, 3)
				obs2 := testObservable[int](ctx, 10, 20, 30)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.HasItems[int]{
					Expected: []int{11, 22, 33},
				})
			})
		})

		When("source observer longer than other", func() {
			It("🧪 should: omit zip from trailing items", func() {
				// rxgo: Test_Observable_ZipFromObservable_DifferentLength1
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2, 3)
				obs2 := testObservable[int](ctx, 10, 20)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.HasItems[int]{
					Expected: []int{11, 22},
				})
			})
		})

		When("source observer shorter than other", func() {
			It("🧪 should: omit zip from trailing items", func() {
				// rxgo: Test_Observable_ZipFromObservable_DifferentLength2
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs1 := testObservable[int](ctx, 1, 2)
				obs2 := testObservable[int](ctx, 10, 20, 30)
				zipper := func(_ context.Context, a, b rx.Item[int]) (int, error) {
					return a.V + b.V, nil
				}
				zip := obs1.ZipFromIterable(obs2, zipper)
				rx.Assert(ctx, zip, rx.HasItems[int]{
					Expected: []int{11, 22},
				})
			})
		})
	})
})
