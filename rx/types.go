package rx

import (
	"context"
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

	// Comparator defines a func that returns an int:
	// - 0 if two elements are equals
	// - A negative value if the first argument is less than the second
	// - A positive value if the first argument is greater than the second
	Comparator[T any] func(T, T) int
	// DistributionFunc used by GroupBy
	DistributionFunc[T any] func(Item[T]) int

	// DistributionFunc used by GroupByDynamic
	DynamicDistributionFunc[T any] func(Item[T]) string

	// InitLimit defines a function to be used with Min and Max operators that defines
	// a limit initialiser, that is to say, for Max we need to initialise the internal
	// maximum reference point to be minimum value for type T and the reverse for the
	// Min operator.
	InitLimit[T any] func() T
	// IsZero determines whether the value T is zero
	IsZero[T any] func(T) bool
	// ItemToObservable defines a function that computes an observable from an item.
	ItemToObservable[T any] func(Item[T]) Observable[T]
	// ErrorToObservable defines a function that transforms an observable from an error.
	ErrorToObservable[T any] func(error) Observable[T]
	// Func defines a function that computes a value from an input value.
	Func[T any] func(context.Context, T) (T, error)
	// Func2 defines a function that computes a value from two input values.
	Func2[T any] func(context.Context, T, T) (T, error)
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
	// Supplier defines a function that supplies a result from nothing.
	Supplier[T any] func(ctx context.Context) Item[T]
	// Disposed is a notification channel indicating when an Observable is closed.
	Disposed <-chan struct{}
	// Disposable is a function to be called in order to dispose a subscription.
	Disposable context.CancelFunc

	// NextFunc handles a next item in a stream.
	NextFunc[T any] func(Item[T])
	// ErrFunc handles an error in a stream.
	ErrFunc func(error)
	// CompletedFunc handles the end of a stream.
	CompletedFunc func()
)

// BackPressureStrategy is the back-pressure strategy type.
type BackPressureStrategy uint32

const (
	// Block blocks until the channel is available.
	Block BackPressureStrategy = iota
	// Drop drops the message.
	Drop
)

// OnErrorStrategy is the Observable error strategy.
type OnErrorStrategy uint32

const (
	// StopOnError is the default error strategy.
	// An operator will stop processing items on error.
	StopOnError OnErrorStrategy = iota
	// ContinueOnError means an operator will continue processing items after an error.
	ContinueOnError
)

// ObservationStrategy defines the strategy to consume from an Observable.
type ObservationStrategy uint32

const (
	// Lazy is the default observation strategy, when an Observer subscribes.
	Lazy ObservationStrategy = iota
	// Eager means consuming as soon as the Observable is created.
	Eager
)
