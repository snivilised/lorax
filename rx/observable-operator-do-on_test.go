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

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

var _ = Describe("Observable operator", func() {
	Context("DoOnComplete", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnCompleted_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				called := false
				<-testObservable[int](ctx, 1, 2, 3).DoOnCompleted(func() {
					called = true
				})

				Expect(called).To(BeTrue())
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnCompleted_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					called := false
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnCompleted(func() {
						called = true
					})

					Expect(called).To(BeTrue())
				})
			})
		})
	})

	Context("DoOnError", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnError_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var got error
				<-testObservable[int](ctx, 1, 2, 3).DoOnError(func(err error) {
					got = err
				})

				Expect(got).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnError_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					var got error
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnError(func(err error) {
						got = err
					})

					Expect(got).To(Equal(errFoo))
				})
			})
		})
	})

	Context("DoOnNext", func() {
		When("no error", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_DoOnNext_NoError
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				s := make([]int, 0)
				<-testObservable[int](ctx, 1, 2, 3).DoOnNext(func(item rx.Item[int]) {
					s = append(s, item.V)
				})

				Expect(s).To(ContainElements([]int{1, 2, 3}))
			})
		})

		Context("Errors", func() {
			When("error", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_DoOnNext_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					s := make([]int, 0)
					<-testObservable[int](ctx, 1, errFoo, 3).DoOnNext(func(item rx.Item[int]) {
						s = append(s, item.V)
					})

					Expect(s).To(ContainElements([]int{1}))
				})
			})
		})
	})
})
