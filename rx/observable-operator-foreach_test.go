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
	Context("ForEach", func() {
		When("principle", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_ForEach_Done
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var gotErr error
				count := 0
				done := make(chan struct{})
				obs := testObservable[int](ctx, 1, 2, 3)
				obs.ForEach(func(i rx.Item[int]) {
					count += i.V
				}, func(err error) {
					gotErr = err
					done <- struct{}{}
				}, func() {
					done <- struct{}{}
				})

				// We avoid using the assertion API on purpose
				<-done

				Expect(count).To(Equal(6))
				Expect(gotErr).To(Succeed())
			})
		})

		Context("Errors", func() {
			When("foo", func() {
				It("🧪 should: ", func() {
					// rxgo: Test_Observable_ForEach_Error
					defer leaktest.Check(GinkgoT())()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					count := 0
					var gotErr error
					done := make(chan struct{})

					obs := testObservable[int](ctx, 1, 2, 3, errFoo)
					obs.ForEach(func(i rx.Item[int]) {
						count += i.V
					}, func(err error) {
						gotErr = err
						select {
						case <-ctx.Done():
							return
						case done <- struct{}{}:
						}
					}, func() {
						select {
						case <-ctx.Done():
							return
						case done <- struct{}{}:
						}
					}, rx.WithContext[int](ctx))

					// We avoid using the assertion API on purpose
					<-done

					Expect(count).To(Equal(6))
					Expect(gotErr).To(MatchError(errFoo))
				})
			})
		})
	})
})
