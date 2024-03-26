package rx

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:revive,stylecheck // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive,stylecheck // gomega ok
)

// AssertPredicate is a custom predicate based on the items.
type AssertPredicate[I any] func(items []I) error

// RxAssert lists the Observable assertions.
type RxAssert[I any] interface { //nolint:revive // foo
	apply(*rxAssert[I])
	itemsToBeChecked() (bool, []I)
	itemsNoOrderedToBeChecked() (bool, []I)
	noItemsToBeChecked() bool
	someItemsToBeChecked() bool
	raisedErrorToBeChecked() (bool, error)
	raisedErrorsToBeChecked() (bool, []error)
	raisedAnErrorToBeChecked() (bool, error)
	notRaisedErrorToBeChecked() bool
	itemToBeChecked() (bool, I)
	noItemToBeChecked() (bool, I)
	customPredicatesToBeChecked() (bool, []AssertPredicate[I])
}

type rxAssert[I any] struct {
	f                       func(*rxAssert[I])
	checkHasItems           bool
	checkHasNoItems         bool
	checkHasSomeItems       bool
	items                   []I
	checkHasItemsNoOrder    bool
	itemsNoOrder            []I
	checkHasRaisedError     bool
	err                     error
	checkHasRaisedErrors    bool
	errs                    []error
	checkHasRaisedAnError   bool
	checkHasNotRaisedError  bool
	checkHasItem            bool
	item                    I
	checkHasNoItem          bool
	checkHasCustomPredicate bool
	customPredicates        []AssertPredicate[I]
}

func (ass *rxAssert[I]) apply(do *rxAssert[I]) {
	ass.f(do)
}

func (ass *rxAssert[I]) itemsToBeChecked() (b bool, i []I) {
	return ass.checkHasItems, ass.items
}

func (ass *rxAssert[I]) itemsNoOrderedToBeChecked() (b bool, i []I) {
	return ass.checkHasItemsNoOrder, ass.itemsNoOrder
}

func (ass *rxAssert[I]) noItemsToBeChecked() bool {
	return ass.checkHasNoItems
}

func (ass *rxAssert[I]) someItemsToBeChecked() bool {
	return ass.checkHasSomeItems
}
func (ass *rxAssert[I]) raisedErrorToBeChecked() (bool, error) {
	return ass.checkHasRaisedError, ass.err
}

func (ass *rxAssert[I]) raisedErrorsToBeChecked() (bool, []error) {
	return ass.checkHasRaisedErrors, ass.errs
}

func (ass *rxAssert[I]) raisedAnErrorToBeChecked() (bool, error) {
	return ass.checkHasRaisedAnError, ass.err
}

func (ass *rxAssert[I]) notRaisedErrorToBeChecked() bool {
	return ass.checkHasNotRaisedError
}

func (ass *rxAssert[I]) itemToBeChecked() (b bool, i I) {
	return ass.checkHasItem, ass.item
}

func (ass *rxAssert[I]) noItemToBeChecked() (b bool, i I) {
	return ass.checkHasNoItem, ass.item
}

func (ass *rxAssert[I]) customPredicatesToBeChecked() (bool, []AssertPredicate[I]) {
	return ass.checkHasCustomPredicate, ass.customPredicates
}

func newAssertion[I any](f func(*rxAssert[I])) *rxAssert[I] {
	return &rxAssert[I]{
		f: f,
	}
}

func parseAssertions[I any](assertions ...RxAssert[I]) RxAssert[I] {
	ass := new(rxAssert[I])

	for _, assertion := range assertions {
		assertion.apply(ass)
	}

	return ass
}

func Assert[I any](ctx context.Context, iterable Iterable[I], assertions ...RxAssert[I]) {
	ass := parseAssertions(assertions...)
	got := make([]I, 0)
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

func HasItems[I any](expectedItems []I) RxAssert[I] {
	return newAssertion(func(ra *rxAssert[I]) {
		ra.checkHasItems = true
		ra.items = expectedItems
	})
}

// HasItem checks if a single or optional single has a specific item.
func HasItem[I any](i I) RxAssert[I] {
	return newAssertion(func(a *rxAssert[I]) {
		a.checkHasItem = true
		a.item = i
	})
}

// IsNotEmpty checks that the observable produces some items.
func IsNotEmpty[I any]() RxAssert[I] {
	return newAssertion(func(a *rxAssert[I]) {
		a.checkHasSomeItems = true
	})
}

// IsEmpty checks that the observable has not produce any item.
func IsEmpty[I any]() RxAssert[I] {
	return newAssertion(func(a *rxAssert[I]) {
		a.checkHasNoItems = true
	})
}

func HasError[I any](err error) RxAssert[I] {
	return newAssertion(func(a *rxAssert[I]) {
		a.checkHasRaisedError = true
		a.err = err
	})
}

// HasAnError checks that the observable has produce an error.
func HasAnError[I any]() RxAssert[I] {
	return newAssertion(func(a *rxAssert[I]) {
		a.checkHasRaisedAnError = true
	})
}

func HasNoError[I any]() RxAssert[I] {
	return newAssertion(func(ra *rxAssert[I]) {
		ra.checkHasNotRaisedError = true
	})
}
