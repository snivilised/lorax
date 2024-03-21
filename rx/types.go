package rx

import "context"

type (
	operatorOptions[I any] struct {
		stop          func()
		resetIterable func(Iterable[I])
	}

	// Comparator defines a func that returns an int:
	// - 0 if two elements are equals
	// - A negative value if the first argument is less than the second
	// - A positive value if the first argument is greater than the second
	Comparator[I any] func(I, I) int
	// ItemToObservable defines a function that computes an observable from an item.
	ItemToObservable[I any] func(Item[I]) Observable[I]
	// ErrorToObservable defines a function that transforms an observable from an error.
	ErrorToObservable[I any] func(error) Observable[I]
	// Func defines a function that computes a value from an input value.
	Func[I any] func(context.Context, I) (I, error)
	// Func2 defines a function that computes a value from two input values.
	Func2[I any] func(context.Context, I, I) (I, error)
	// FuncN defines a function that computes a value from N input values.
	FuncN[I any] func(...I) I
	// ErrorFunc defines a function that computes a value from an error.
	ErrorFunc[I any] func(error) I
	// Predicate defines a func that returns a bool from an input value.
	Predicate[I any] func(I) bool
	// Marshaller defines a marshaller type (ItemValue[I] to []byte).
	Marshaller[I any] func(I) ([]byte, error)
	// Unmarshaller defines an unmarshaller type ([]byte to interface).
	Unmarshaller[I any] func([]byte, I) error
	// Producer defines a producer implementation.
	Producer[I any] func(ctx context.Context, next chan<- Item[I])
	// Supplier defines a function that supplies a result from nothing.
	Supplier[I any] func(ctx context.Context) Item[I]
	// Disposed is a notification channel indicating when an Observable is closed.
	Disposed <-chan struct{}
	// Disposable is a function to be called in order to dispose a subscription.
	Disposable context.CancelFunc

	// NextFunc handles a next item in a stream.
	NextFunc[I any] func(I)
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
