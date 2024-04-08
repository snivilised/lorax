package rx_test

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
