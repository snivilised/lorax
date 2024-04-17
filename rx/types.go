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

	"golang.org/x/exp/constraints"
)

type (
	operatorOptions[T any] struct {
		stop          func()
		resetIterable func(Iterable[T])
	}

	// Calculator defines numeric operations for T
	Calculator[T any] interface {
		Add(T, T) T
		Div(T, T) T
		Inc(T) T
		IsZero(T) bool
		Zero() T
	}

	// CalculatorItem defines numeric operations for T
	CalculatorItem[T any] interface {
		Add(Item[T], Item[T]) Item[T]
		Div(Item[T], Item[T]) Item[T]
		Inc(Item[T]) Item[T]
		IsZero(Item[T]) bool
		Zero() Item[T]
	}

	// Comparator defines a func that returns an int:
	// - 0 if two elements are equals
	// - A negative value if the first argument is less than the second
	// - A positive value if the first argument is greater than the second
	Comparator[T any] func(Item[T], Item[T]) int
	// DistributionFunc used by GroupBy
	DistributionFunc[T any] func(Item[T]) int

	// DistributionFunc used by GroupByDynamic
	DynamicDistributionFunc[T any] func(Item[T]) string

	// InitLimit defines a function to be used with Min and Max operators that defines
	// a limit initialiser, that is to say, for Max we need to initialise the internal
	// maximum reference point to be minimum value for type T and the reverse for the
	// Min operator.
	InitLimit[T any] func() Item[T]
	// IsZero determines whether the value T is zero
	IsZero[T any] func(T) bool
	// ItemToObservable defines a function that computes an observable from an item.
	ItemToObservable[T any] func(Item[T]) Observable[T]
	// ErrorToObservable defines a function that transforms an observable from an error.
	ErrorToObservable[T any] func(error) Observable[T]
	// Func defines a function that computes a value from an input value.
	Func[T any] func(context.Context, T) (T, error)
	// Func2 defines a function that computes a value from two input values.
	Func2[T any] func(context.Context, Item[T], Item[T]) (T, error)
	// FuncIntM defines a function that's specialised for Map
	// To solve the problem of being able to map values across different
	// types, the FuncIntM type will be modified to take an extra type
	// parameter 'O' which represents the 'Other' type, ie we map from
	// a value of type 'T' to a value of type 'O' (Func[T, O any]). With
	// this in place, we should be able to define a pipeline that starts
	// off with values of type T, and end up with values of type O via a
	// Map operator. We'll have to make sure that any intermediate
	// channels can be appropriately defined. If we want to map between
	// values within the same type, we need to use a separate definition,
	// and in fact the original definition Func, should suffice.
	//
	// The problem with Map is that it introduces a new Type. Perhaps we need
	// a separate observer interface that can bridge from T to O. We can't
	// introduce the type O, because that would have a horrendous cascading
	// impact on every type, when most operations would not need this type O.
	// With generics, Map is a very awkward operator that needs special attention.
	// In the short term, what we can say is that the base functionality only
	// allows mapping to different values within the same type.
	// FuncIntM[T any] func(context.Context, int) (int, error)
	// FuncN defines a function that computes a value from N input values.
	FuncN[T any] func(...T) T
	// ErrorFunc defines a function that computes a value from an error.
	ErrorFunc[T any] func(error) T
	// Predicate defines a func that returns a bool from an input item.
	Predicate[T any] func(Item[T]) bool
	// Marshaller defines a marshaller type (ItemValue[T] to []byte).
	Marshaller[T any] func(T) ([]byte, error)
	// Unmarshaller defines an unmarshaller type ([]byte to interface).
	Unmarshaller[T any] func([]byte, T) error
	// Producer defines a producer implementation.
	Producer[T any] func(ctx context.Context, next chan<- Item[T])
	// ShouldRetryFunc as used by Retry operator
	ShouldRetryFunc func(error) bool
	// Supplier defines a function that supplies a result from nothing.
	Supplier[T any] func(ctx context.Context) Item[T]
	// Disposed is a notification channel indicating when an Observable is closed.
	Disposed <-chan struct{}
	// Disposable is a function to be called in order to dispose a subscription.
	Disposable context.CancelFunc
	// NextFunc handles a next item in a stream.
	NextFunc[T any] func(Item[T])
	// NumVal is an integer value used by Item.N and Range
	NumVal = int
	// ErrFunc handles an error in a stream.
	ErrFunc func(error)
	// CompletedFunc handles the end of a stream.
	CompletedFunc func()
	// Numeric defines a constraint that targets scalar types for whom numeric
	// operators are natively defined.
	Numeric interface {
		constraints.Integer | constraints.Signed | constraints.Unsigned | constraints.Float
	}
	// WhilstFunc condition function as used by Range
	WhilstFunc[T any] func(current T) bool
	// RangeIterator allows the client defines how the Range operator emits derived
	// items.
	RangeIterator[T any] interface {
		Init() error
		// Start should return the initial index value
		Start() (*T, error)
		// Step is used by Increment and defines the size of increment for each iteration
		Step() T
		// Increment increments the index value
		Increment(index *T) T
		// Plus(index, by T)
		// While defines a condition that must be true for the loop to
		// continue iterating.
		While(current T) bool
	}

	NominatedField[T any, O Numeric] interface {
		Field() O
		Inc(index *T, by T) *T
		Index(int) *T
	}

	RangeIteratorNF[T NominatedField[T, O], O Numeric] interface {
		Init() error
		// Start should return the initial index value
		Start() (*T, error)
		// Step is used by Increment and defines the size of increment for each iteration
		Step() O
		// Increment returns a pointer to a new instance of with incremented index value
		Increment(index *T) *T
		// While defines a condition that must be true for the loop to
		// continue iterating.
		While(current T) bool
	}
)
