package rxa

func HasItems[I any](expectedItems []I) RxAssert[I] {
	return newAssertion(func(ra *rxAssert[I]) {
		ra.checkHasItems = true
		ra.items = expectedItems
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
