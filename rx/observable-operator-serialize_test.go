package rx_test

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"context"
	"fmt"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	"github.com/snivilised/lorax/rx"
)

type message struct {
	id int
}

var _ = Describe("Observable operator", func() {
	Context("Serialize", func() {
		When("struct", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_Struct
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[message](ctx,
					message{3}, message{5}, message{1}, message{2}, message{4},
				).Serialize(1, func(i interface{}) int {
					return i.(message).id
				})
				rx.Assert(ctx, obs,
					rx.HasItems[message]{
						Expected: []message{
							{1}, {2}, {3}, {4}, {5},
						},
					},
				)
			})
		})

		When("duplicates", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_Struct
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[int](ctx, 1, 3, 2, 6, 4, 5).
					Serialize(1, func(i interface{}) int {
						return i.(int)
					})
				rx.Assert(ctx, obs, rx.HasItems[int]{
					Expected: []int{1, 2, 3, 4, 5, 6},
				})
			})
		})

		When("loop", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_Loop
				defer leaktest.Check(GinkgoT())()

				idx := 0
				<-rx.Range[int](1, 10000).
					Serialize(0, func(i any) int {
						return i.(int)
					}).
					Map(func(_ context.Context, i int) (int, error) {
						return i, nil
					}, rx.WithCPUPool[int]()).
					DoOnNext(func(it rx.Item[int]) {
						v := it.V

						if v != idx {
							Fail(fmt.Sprintf("not sequential, expected=%d, got=%d", idx, v))
						}
						idx++
					})
			})
		})

		When("Different from", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_DifferentFrom
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[message](ctx,
					message{13}, message{15}, message{11}, message{12}, message{14},
				).Serialize(11, func(i interface{}) int {
					return i.(message).id
				})
				rx.Assert(ctx, obs,
					rx.HasItems[message]{
						Expected: []message{
							{11}, {12}, {13}, {14}, {15},
						},
					},
				)
			})
		})

		When("Context canceled", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_ContextCanceled
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()

				obs := rx.Never[message]().Serialize(1, func(i interface{}) int {
					return i.(message).id
				}, rx.WithContext[message](ctx))
				rx.Assert(ctx, obs,
					rx.IsEmpty[message]{},
					rx.HasNoError[message]{},
				)
			})
		})

		When("empty", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Serialize_Empty
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				obs := testObservable[message](ctx,
					message{3}, message{5}, message{7}, message{2}, message{4},
				).Serialize(1, func(i interface{}) int {
					return i.(message).id
				},
				)
				rx.Assert(ctx, obs, rx.IsEmpty[message]{})
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_Serialize_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					obs := testObservable[message](ctx,
						message{3}, message{1}, errFoo, message{2}, message{4},
					).Serialize(1, func(i interface{}) int {
						return i.(message).id
					})
					rx.Assert(ctx, obs,
						rx.HasItems[message]{
							Expected: []message{{1}},
						},
						rx.HasError[message]{
							Expected: []error{errFoo},
						},
					)
				})
			})
		})
	})
})
