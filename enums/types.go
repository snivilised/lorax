package enums

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
