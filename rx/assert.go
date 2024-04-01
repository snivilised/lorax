package rx

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive,stylecheck // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive,stylecheck // gomega ok
)

// AssertPredicate is a custom predicate based on the items.
type AssertPredicate[T any] func(items []T) error

// RxAssert lists the Observable assertions.
type RxAssert[T any] interface { //nolint:revive // foo
	apply(*rxAssert[T])
	itemsToBeChecked() (bool, []T)
	itemsNoOrderedToBeChecked() (bool, []T)
	noItemsToBeChecked() bool
	someItemsToBeChecked() bool
	raisedErrorToBeChecked() (bool, error)
	raisedErrorsToBeChecked() (bool, []error)
	raisedAnErrorToBeChecked() (bool, error)
	notRaisedErrorToBeChecked() bool
	itemToBeChecked() (bool, T)
	noItemToBeChecked() (bool, T)
	customPredicatesToBeChecked() (bool, []AssertPredicate[T])
}

type rxAssert[T any] struct {
	f                       func(*rxAssert[T])
	checkHasItems           bool
	checkHasNoItems         bool
	checkHasSomeItems       bool
	items                   []T
	checkHasItemsNoOrder    bool
	itemsNoOrder            []T
	checkHasRaisedError     bool
	err                     error
	checkHasRaisedErrors    bool
	errs                    []error
	checkHasRaisedAnError   bool
	checkHasNotRaisedError  bool
	checkHasItem            bool
	item                    T
	checkHasNoItem          bool
	checkHasCustomPredicate bool
	customPredicates        []AssertPredicate[T]
}

func (ass *rxAssert[T]) apply(do *rxAssert[T]) {
	ass.f(do)
}

func (ass *rxAssert[T]) itemsToBeChecked() (b bool, i []T) {
	return ass.checkHasItems, ass.items
}

func (ass *rxAssert[T]) itemsNoOrderedToBeChecked() (b bool, i []T) {
	return ass.checkHasItemsNoOrder, ass.itemsNoOrder
}

func (ass *rxAssert[T]) noItemsToBeChecked() bool {
	return ass.checkHasNoItems
}

func (ass *rxAssert[T]) someItemsToBeChecked() bool {
	return ass.checkHasSomeItems
}
func (ass *rxAssert[T]) raisedErrorToBeChecked() (bool, error) {
	return ass.checkHasRaisedError, ass.err
}

func (ass *rxAssert[T]) raisedErrorsToBeChecked() (bool, []error) {
	return ass.checkHasRaisedErrors, ass.errs
}

func (ass *rxAssert[T]) raisedAnErrorToBeChecked() (bool, error) {
	return ass.checkHasRaisedAnError, ass.err
}

func (ass *rxAssert[T]) notRaisedErrorToBeChecked() bool {
	return ass.checkHasNotRaisedError
}

func (ass *rxAssert[T]) itemToBeChecked() (b bool, i T) {
	return ass.checkHasItem, ass.item
}

func (ass *rxAssert[T]) noItemToBeChecked() (b bool, i T) {
	return ass.checkHasNoItem, ass.item
}

func (ass *rxAssert[T]) customPredicatesToBeChecked() (bool, []AssertPredicate[T]) {
	return ass.checkHasCustomPredicate, ass.customPredicates
}

func newAssertion[T any](f func(*rxAssert[T])) *rxAssert[T] {
	return &rxAssert[T]{
		f: f,
	}
}

func parseAssertions[T any](assertions ...RxAssert[T]) RxAssert[T] {
	ass := new(rxAssert[T])

	for _, assertion := range assertions {
		assertion.apply(ass)
	}

	return ass
}

func Assert[T any](ctx context.Context, iterable Iterable[T], assertions ...RxAssert[T]) {
	ass := parseAssertions(assertions...)
	got := make([]T, 0)
	errs := make([]error, 0)
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

			// TODO: needs to accommodate item.N, ie the numeric aux value
			// and also should be modified to support all the other
			// new ways of interpreting an item (Ch, Tick, Tv), possibly
			// with new assertions, ie: HasCh, HasTick, HasTv.
			//
			if item.IsError() {
				errs = append(errs, item.E)
			} else {
				got = append(got, item.V)
			}
		}
	}

	if checked, predicates := ass.customPredicatesToBeChecked(); checked {
		for _, predicate := range predicates {
			err := predicate(got)
			if err != nil {
				Fail(err.Error())
			}
		}
	}

	if checkHasItems, expectedItems := ass.itemsToBeChecked(); checkHasItems {
		Expect(got).To(ContainElements(expectedItems))
	}

	if checkHasItemsNoOrder, itemsNoOrder := ass.itemsNoOrderedToBeChecked(); checkHasItemsNoOrder {
		m := make(map[interface{}]interface{})
		for _, v := range itemsNoOrder {
			m[v] = nil
		}

		for _, v := range got {
			delete(m, v)
		}

		if len(m) != 0 {
			Fail(fmt.Sprintf("missing elements: '%v'", got))
		}
	}

	if checkHasItem, value := ass.itemToBeChecked(); checkHasItem {
		length := len(got)
		if length != 1 {
			Fail(fmt.Sprintf("wrong number of items, expected 1, got %d", length))
		}

		if length > 0 {
			Expect(got[0]).To(Equal(value))
		}
	}

	if ass.noItemsToBeChecked() {
		Expect(got).To(BeEmpty())
	}

	if ass.someItemsToBeChecked() {
		Expect(got).NotTo(BeEmpty())
	}

	if checkHasRaisedError, expectedError := ass.raisedErrorToBeChecked(); checkHasRaisedError {
		if expectedError == nil {
			Expect(errs).To(BeEmpty())
		} else {
			length := len(errs)

			if length == 0 {
				Fail(fmt.Sprintf("no error raised; expected: %v", expectedError))
			}

			if length > 0 {
				Expect(errs[0]).Error().To(Equal(expectedError))
			}
		}
	}

	if checkHasRaisedErrors, expectedErrors := ass.raisedErrorsToBeChecked(); checkHasRaisedErrors {
		Expect(errs).To(ContainElements(expectedErrors))
	}

	if checkHasRaisedAnError, expectedError := ass.raisedAnErrorToBeChecked(); checkHasRaisedAnError {
		Expect(expectedError).Error().To(BeNil()) // this might not be right
	}

	if ass.notRaisedErrorToBeChecked() {
		Expect(errs).To(BeEmpty())
	}
}

func HasItems[T any](expectedItems []T) RxAssert[T] {
	return newAssertion(func(ra *rxAssert[T]) {
		ra.checkHasItems = true
		ra.items = expectedItems
	})
}

// HasItem checks if a single or optional single has a specific item.
func HasItem[T any](i T) RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasItem = true
		a.item = i
	})
}

// HasItemsNoOrder checks that an observable produces the corresponding items regardless of the order.
func HasItemsNoOrder[T any](items ...T) RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasItemsNoOrder = true
		a.itemsNoOrder = items
	})
}

// IsNotEmpty checks that the observable produces some items.
func IsNotEmpty[T any]() RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasSomeItems = true
	})
}

// IsEmpty checks that the observable has not produce any item.
func IsEmpty[T any]() RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasNoItems = true
	})
}

func HasError[T any](err error) RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasRaisedError = true
		a.err = err
	})
}

// HasAnError checks that the observable has produce an error.
func HasAnError[T any]() RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		a.checkHasRaisedAnError = true
	})
}

func HasNoError[T any]() RxAssert[T] {
	return newAssertion(func(ra *rxAssert[T]) {
		ra.checkHasNotRaisedError = true
	})
}

// CustomPredicate checks a custom predicate.
func CustomPredicate[T any](predicate AssertPredicate[T]) RxAssert[T] {
	return newAssertion(func(a *rxAssert[T]) {
		if !a.checkHasCustomPredicate {
			a.checkHasCustomPredicate = true
			a.customPredicates = make([]AssertPredicate[T], 0)
		}

		a.customPredicates = append(a.customPredicates, predicate)
	})
}
