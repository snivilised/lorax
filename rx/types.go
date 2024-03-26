package rx

import "context"

type (
	operatorOptions[T any] struct {
		stop          func()
		resetIterable func(Iterable[T])
	}

	// Comparator defines a func that returns an int:
	// - 0 if two elements are equals
	// - A negative value if the first argument is less than the second
	// - A positive value if the first argument is greater than the second
	Comparator[T any] func(T, T) int
	// ItemToObservable defines a function that computes an observable from an item.
	ItemToObservable[T any] func(Item[T]) Observable[T]
	// ErrorToObservable defines a function that transforms an observable from an error.
	ErrorToObservable[T any] func(error) Observable[T]
	// Func defines a function that computes a value from an input value.
	Func[T any] func(context.Context, T) (T, error)
	// Func2 defines a function that computes a value from two input values.
	Func2[T any] func(context.Context, T, T) (T, error)
	// FuncN defines a function that computes a value from N input values.
	FuncN[T any] func(...T) T
	// ErrorFunc defines a function that computes a value from an error.
	ErrorFunc[T any] func(error) T
	// Predicate defines a func that returns a bool from an input value.
	Predicate[T any] func(T) bool
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
	NextFunc[T any] func(T)
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
