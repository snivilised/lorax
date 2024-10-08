package rx

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

	. "github.com/onsi/ginkgo/v2" //nolint:revive,stylecheck // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive,stylecheck // gomega ok
	"github.com/snivilised/lorax/enums"
)

// AssertPredicate is a custom predicate based on the items.
type AssertPredicate[T any] func(actual AssertResources[T]) error

func reason(s string) string {
	return fmt.Sprintf("🔥🔥🔥 %v", s)
}

type AssertResources[T any] interface {
	Values() []T
	Numbers() []int
	Errors() []error
	Booleans() []bool
	Ticks() []NumVal
	TickValues() []NumVal
	Opaques() []any
}

type actualResources[T any] struct {
	values     []T
	numbers    []int
	errors     []error
	booleans   []bool
	ticks      []NumVal
	tickValues []NumVal
	opaques    []any
}

func (r *actualResources[T]) Values() []T {
	return r.values
}

func (r *actualResources[T]) Numbers() []int {
	return r.numbers
}

func (r *actualResources[T]) Errors() []error {
	return r.errors
}

func (r *actualResources[T]) Booleans() []bool {
	return r.booleans
}

func (r *actualResources[T]) Ticks() []NumVal {
	return r.ticks
}

func (r *actualResources[T]) TickValues() []NumVal {
	return r.tickValues
}

func (r *actualResources[T]) Opaques() []any {
	return r.opaques
}

func Assert[T any](ctx context.Context, iterable Iterable[T], asserters ...Asserter[T]) {
	resources := assertObserver(ctx, iterable)

	for _, a := range asserters {
		a.Check(resources)
	}
}

func assertObserver[T any](ctx context.Context, iterable Iterable[T]) *actualResources[T] {
	resources := &actualResources[T]{
		values:     make([]T, 0),
		numbers:    make([]NumVal, 0),
		errors:     make([]error, 0),
		booleans:   make([]bool, 0),
		ticks:      make([]NumVal, 0),
		tickValues: make([]NumVal, 0),
		opaques:    make([]any, 0),
	}

	observe := iterable.Observe()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case item, ok := <-observe:
			if !ok {
				break loop
			}

			switch item.disc {
			case enums.ItemDiscNative:
				resources.values = append(resources.values, item.V)

			case enums.ItemDiscBoolean:
				resources.booleans = append(resources.booleans, item.Bool())

			case enums.ItemDiscWChan:
				Fail(fmt.Sprintf("NOT-IMPL yet for channel for: '%v'", item.Desc()))

			case enums.ItemDiscError:
				resources.errors = append(resources.errors, item.E)

			case enums.ItemDiscNumber:
				resources.numbers = append(resources.numbers, item.Num())

			case enums.ItemDiscTick:
				resources.ticks = append(resources.ticks, item.Num())

			case enums.ItemDiscTickValue:
				resources.tickValues = append(resources.tickValues, item.Num())

			case enums.ItemDiscOpaque:
				resources.opaques = append(resources.opaques, item.Opaque())

			case enums.ItemDiscUndefined:
				Fail(fmt.Sprintf("undefined value not handled for: '%v'", item.Desc()))

			default:
				Fail(fmt.Sprintf("value type not handled for: '%v'", item.Desc()))
			}
		}
	}

	return resources
}

// Asserter
type Asserter[T any] interface {
	Check(actual AssertResources[T])
}

type AssertFunc[T any] func(actual AssertResources[T])

func (f AssertFunc[T]) Check(actual AssertResources[T]) {
	f(actual)
}

// ContainItems
type ContainItems[T any] struct {
	Expected []T
}

// HasItems checks if an observable contains expected items
func (a ContainItems[T]) Check(actual AssertResources[T]) {
	Expect(actual.Values()).To(ContainElements(a.Expected), reason("HasItems"))
}

// HasItems
type HasItems[T any] struct {
	Expected []T
}

// HasItems checks if an observable has an exact set of items.
func (a HasItems[T]) Check(actual AssertResources[T]) {
	Expect(actual.Values()).To(HaveExactElements(a.Expected), reason("HasItems"))
}

// HasItem
type HasItem[T any] struct {
	Expected T
}

// HasItem checks if a single or optional single has a specific item.
func (a HasItem[T]) Check(actual AssertResources[T]) {
	values := actual.Values()
	length := len(values)

	if length != 1 {
		Fail(reason(fmt.Sprintf("HasItem: wrong number of items, expected 1, got %d", length)))
	}

	if length > 0 {
		Expect(values[0]).To(Equal(a.Expected), reason("HasItem"))
	}
}

// HasNumbers
type HasNumbers[T any] struct {
	Expected []T
}

// HasNumbers checks if an observable has an exact set of numeric items.
func (a HasNumbers[T]) Check(actual AssertResources[T]) {
	Expect(actual.Numbers()).To(ContainElements(a.Expected), reason("HasNumbers"))
}

// HasNumber
type HasNumber[T any] struct {
	Expected int
}

// HasNumber checks if a single or optional single has a specific numeric item.
func (a HasNumber[T]) Check(actual AssertResources[T]) {
	values := actual.Numbers()
	length := len(values)

	if length != 1 {
		Fail(reason(fmt.Sprintf("HasNumber: wrong number of items, expected 1, got %d", length)))
	}

	if length > 0 {
		Expect(values[0]).To(Equal(a.Expected), reason("HasNumber"))
	}
}

// HasItemsNoOrder
type HasItemsNoOrder[T any] struct {
	Expected []T
}

// Check ensures that an observable produces the corresponding items regardless of the order.
func (a HasItemsNoOrder[T]) Check(actual AssertResources[T]) {
	values := actual.Values()
	m := make(map[any]any)

	for _, v := range a.Expected {
		m[v] = nil
	}

	for _, v := range values {
		delete(m, v)
	}

	if len(m) != 0 {
		Fail(reason(fmt.Sprintf("HasItemsNoOrder: missing elements: '%v'", values)))
	}
}

// HasNumbersNoOrder
type HasNumbersNoOrder[T any] struct {
	Expected []T
}

// Check ensures that an observable produces the corresponding numbers regardless of the order.
func (a HasNumbersNoOrder[T]) Check(actual AssertResources[T]) {
	values := actual.Numbers()
	m := make(map[any]any)

	for _, v := range a.Expected {
		m[v] = nil
	}

	for _, v := range values {
		delete(m, v)
	}

	if len(m) != 0 {
		Fail(reason(fmt.Sprintf("HasNumbersNoOrder: missing elements: '%v'", values)))
	}
}

// IsNotEmpty
type IsNotEmpty[T any] struct {
}

func (a IsNotEmpty[T]) Check(actual AssertResources[T]) {
	// TODO: what about numeric items? What actually does NotEmpty mean?
	Expect(actual.Values()).NotTo(BeEmpty(), reason("IsNotEmpty"))
}

// IsEmpty
type IsEmpty[T any] struct {
}

func (a IsEmpty[T]) Check(actual AssertResources[T]) {
	// TODO: what about numeric items? What actually does NotEmpty mean?
	Expect(actual.Values()).To(BeEmpty(), reason("IsEmpty"))
}

// HasError
type HasError[T any] struct {
	Expected []error
}

func (a HasError[T]) Check(actual AssertResources[T]) {
	errors := actual.Errors()

	if a.Expected == nil || len(a.Expected) == 0 {
		Expect(errors).To(BeEmpty(), reason("HasError"))

		return
	}

	if len(errors) == 0 {
		Fail(fmt.Sprintf("HasError: no error raised; expected: %v", a.Expected))
	}

	Expect(errors).To(ContainElements(a.Expected), reason("HasError"))
}

// HasAnError
type HasAnError[T any] struct {
	Expected error
}

// Check HasAnError ensures that the observable has produced a specific error.
func (a HasAnError[T]) Check(actual AssertResources[T]) {
	errors := actual.Errors()

	Expect(errors).NotTo(BeEmpty(), reason("HasAnError: no errors occurred"))
}

// HasNoError
type HasNoError[T any] struct {
}

// Check HasNoError ensures that the observable has not produced an error.
func (a HasNoError[T]) Check(actual AssertResources[T]) {
	Expect(actual.Errors()).To(BeEmpty(), reason("HasNoError"))
}

// CustomPredicate
type CustomPredicate[T any] struct {
	Expected AssertPredicate[T]
}

// Check CustomPredicateAssert checks a custom predicate.
func (a CustomPredicate[T]) Check(actual AssertResources[T]) {
	Expect(a.Expected(actual)).To(Succeed(), reason("CustomPredicate"))
}

// HasTrue
type HasTrue[T any] struct {
}

// Check HasTrue checks boolean values contains at least 1 true value
func (a HasTrue[T]) Check(actual AssertResources[T]) {
	values := actual.Booleans()

	if len(values) == 0 {
		Fail("HasTrue: no values found")
	}

	Expect(values).To(ContainElements(true), reason("HasTrue"))
}

// HasFalse
type HasFalse[T any] struct {
}

// Check HasFalse checks boolean values contains at least 1 true false
func (a HasFalse[T]) Check(actual AssertResources[T]) {
	values := actual.Booleans()

	if len(values) == 0 {
		Fail("HasFalse: no values found")
	}

	Expect(values).To(ContainElements(false), reason("HasFalse"))
}

// IsTrue
type IsTrue[T any] struct {
}

// Check IsTrue checks boolean value is true
func (a IsTrue[T]) Check(actual AssertResources[T]) {
	values := actual.Booleans()

	if len(values) == 0 {
		Fail("IsTrue: no value found")
	}

	Expect(values[0]).To(BeTrue(), reason("IsTrue"))
}

// IsFalse
type IsFalse[T any] struct {
}

// Check IsFalse checks boolean value is false
func (a IsFalse[T]) Check(actual AssertResources[T]) {
	values := actual.Booleans()

	if len(values) == 0 {
		Fail("IsFalse: no value found")
	}

	Expect(values[0]).To(BeFalse(), reason("IsFalse"))
}

// HasTicks
type HasTicks[T any] struct {
	Expected []T
}

// HasTickValues checks if an observable has an exact set of numeric items.
func (a HasTicks[T]) Check(actual AssertResources[T]) {
	Expect(actual.Ticks()).To(ContainElements(a.Expected), reason("HasTickValues"))
}

// HasTickCount
type HasTickCount[T any] struct {
	Expected int
}

// HasTickValues checks if an observable has a expect count of tick value items.
func (a HasTickCount[T]) Check(actual AssertResources[T]) {
	Expect(actual.Ticks()).To(HaveLen(a.Expected), reason("HasTickCount"))
}

// HasTickValue
type HasTickValue[T any] struct {
	Expected int
}

// HasTickValue checks if a single or optional single has a specific numeric item.
func (a HasTickValue[T]) Check(actual AssertResources[T]) {
	values := actual.TickValues()
	length := len(values)

	if length != 1 {
		Fail(reason(fmt.Sprintf("HasTickValue: wrong number of items, expected 1, got %d", length)))
	}

	if length > 0 {
		Expect(values[0]).To(Equal(a.Expected), reason("HasTickValue"))
	}
}

// HasTickValueCount
type HasTickValueCount[T any] struct {
	Expected int
}

// HasTickValues checks if an observable has a expect count of tick value items.
func (a HasTickValueCount[T]) Check(actual AssertResources[T]) {
	Expect(actual.TickValues()).To(HaveLen(a.Expected), reason("HasTickValueCount"))
}
