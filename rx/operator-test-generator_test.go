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
	"fmt"

	"github.com/fortytw2/leaktest"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
	"github.com/snivilised/lorax/rx"
)

type generateTE struct {
	name string
}

var _ = Describe("Operator-Test-Generator", func() {
	Context("generator", func() {
		When("JustItem", func() {
			It("🧪 should: get the single value", func() {
				defer leaktest.Check(GinkgoT())()

				Expect(1).To(Equal(1))
				rx.Just("delete-me")
			})
		})

		DescribeTable("operators",
			func(entry *generateTE) {
				path := entry.name

				fmt.Printf("file-path: '%s'\n", path)
			},
			func(entry *generateTE) string {
				return fmt.Sprintf("🧪 ===> should: generate test for operator: '%v'", entry.name)
			},
		)
	})
})

// All
// AverageFloat32 / AverageFloat64
// AverageInt / AverageInt8 / AverageInt16 / AverageInt32 / AverageInt64
// BackOffRetry
// BufferWithCount / BufferWithTime
// Contains
// Range
// Debounce
// DefaultIfEmpty
// Distinct
// DoOnCompleted / DoOnError / DoOnNext
// ElementAt
// Error / Errors (multiple test sections)
// Filter
// Find
// First
// FlatMap
// ForEach
// IgnoreElements
// GroupBy
// Join
// Last
// Map
// Marshal
// Max
// Min
// Observe
// OnErrorResumeNext / OnErrorReturn / OnErrorReturnItem
// Reduce
// Repeat
// Retry
// Run
// Sample
// Scan
// SequenceEqual
// Send
// Serialize
// Skip / SkipLast / SkipWhile
// StartWith
// SumFloat32 / SumFloat64 / SumInt64
// Take / TakeLast / TakeUntil / TakeWhile
// TimeInterval
// TimeStamp
// ToMap / ToMapWithValueSelector
// ToSlice
// UnMarshal
// WindowWithCount / WindowWithTimeOrCount
// ZipFromIterable / ZipFromIterable
