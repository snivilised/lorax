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
	Context("Send", func() {
		When("channel is buffered", func() {
			It("🧪 should: ", func() {
				// rxgo: Test_Observable_Send
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int], 10)
				testObservable[int](ctx, 1, 2, 3, errFoo).Send(ch)
				Expect(rx.Of(1)).To(Equal(<-ch))
				Expect(rx.Of(2)).To(Equal(<-ch))
				Expect(rx.Of(3)).To(Equal(<-ch))
				Expect(rx.Error[int](errFoo)).To(Equal(<-ch))
			})
		})

		When("channel is not buffered", func() {
			It("🧪 should: ", func() {
				defer leaktest.Check(GinkgoT())()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ch := make(chan rx.Item[int])
				testObservable[int](ctx, 1, 2, 3, errFoo).Send(ch)
				Expect(rx.Of(1)).To(Equal(<-ch))
				Expect(rx.Of(2)).To(Equal(<-ch))
				Expect(rx.Of(3)).To(Equal(<-ch))
				Expect(rx.Error[int](errFoo)).To(Equal(<-ch))
			})
		})
	})
})
