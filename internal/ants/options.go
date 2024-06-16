package ants

import (
	"runtime"
	"time"
)

// Option represents the functional option.
type Option func(opts *Options)

// NewOptions creates new options instance with defaults
// applied.
func NewOptions(options ...Option) *Options {
	opts := new(Options)
	combined := withDefaults(options...)

	for _, option := range combined {
		if option != nil { // nil check supports conditional options
			option(opts)
		}
	}

	return opts
}

// If enables options to be conditional. If condition evaluates to true
// then the option is returned, otherwise nil.
func If(condition bool, option Option) Option {
	if condition {
		return option
	}

	return nil
}

func withDefaults(options ...Option) []Option {
	defaults := []Option{
		WithGenerator(&Sequential{
			Format: "ID:%08d",
		}),
		WithSize(uint(runtime.NumCPU())),
	}

	o := make([]Option, 0, len(options)+len(defaults))
	o = append(o,
		defaults...,
	)
	o = append(o, options...)

	return o
}

// Options contains all options which will be applied when instantiating an ants pool.
type Options struct {
	// ExpiryDuration is a period for the scavenger goroutine to clean up those expired workers,
	// the scavenger scans all workers every `ExpiryDuration` and clean up those workers that haven't been
	// used for more than `ExpiryDuration`.
	ExpiryDuration time.Duration

	// PreAlloc indicates whether to make memory pre-allocation when initializing Pool.
	PreAlloc bool

	// Max number of goroutine blocking on pool.Submit.
	// 0 (default value) means no such limit.
	MaxBlockingTasks int

	// When Nonblocking is true, Pool.Submit will never be blocked.
	// ErrPoolOverload will be returned when Pool.Submit cannot be done at once.
	// When Nonblocking is true, MaxBlockingTasks is inoperative.
	Nonblocking bool

	// PanicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	PanicHandler func(interface{})

	// Logger is the customized logger for logging info, if it is not set,
	// default standard logger from log package is used.
	Logger Logger

	// When DisablePurge is true, workers are not purged and are resident.
	DisablePurge bool

	// Size denotes the number of workers in the pool. Defaults
	// to number of CPUs available, if not specified.
	Size uint

	// Generator used to generate job ids.
	Generator IDGenerator

	// Input options
	Input InputOptions

	// Output options
	Output *OutputOptions
}

type InputOptions struct {
	// BufferSize
	BufferSize uint
}

const (
	// MinimumCheckCloseInterval denotes the minimum duration of how long to wait
	// in between successive attempts to check wether the output channel can be
	// closed when the source of the workload indicates no more jobs will be
	// submitted, either by closing the input stream or invoking Conclude on the pool.
	//
	MinimumCheckCloseInterval = time.Millisecond * 10

	// MinimumTimeoutOnSend denotes the minimum duration of how long to allow for
	// when sending output. When this timeout occurs, the worker will send a
	// cancellation request back to the client via the cancellation channel at which
	// point it can cancel the whole worker pool.
	//
	MinimumTimeoutOnSend = time.Millisecond * 10
)

type OutputOptions struct {
	// BufferSize
	BufferSize uint

	// Interval denotes how long to wait in between successive attempts
	// to check wether the output channel can be closed when the source
	// of the workload indicates no more jobs will be submitted, either
	// by closing the input stream or invoking Conclude on the pool.
	//
	CheckCloseInterval time.Duration

	// TimeoutOnSend denotes how long to allow for when sending output.
	// When this timeout occurs, the worker will send a cancellation
	// request back to the client via the cancellation channel at which
	// point it can cancel the whole worker pool.
	//
	TimeoutOnSend time.Duration
}

// WithOptions accepts the whole options config.
func WithOptions(options Options) Option { //nolint:gocritic // heavy options not important
	return func(opts *Options) {
		*opts = options
	}
}

// WithExpiryDuration sets up the interval time of cleaning up goroutines.
func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(opts *Options) {
		opts.ExpiryDuration = expiryDuration
	}
}

// WithPreAlloc indicates whether it should malloc for workers.
func WithPreAlloc(preAlloc bool) Option {
	return func(opts *Options) {
		opts.PreAlloc = preAlloc
	}
}

// WithMaxBlockingTasks sets up the maximum number of goroutines that are
// blocked when it reaches the capacity of pool.
func WithMaxBlockingTasks(maxBlockingTasks int) Option {
	return func(opts *Options) {
		opts.MaxBlockingTasks = maxBlockingTasks
	}
}

// WithNonblocking indicates that pool will return nil when there is no
// available workers.
func WithNonblocking(nonblocking bool) Option {
	return func(opts *Options) {
		opts.Nonblocking = nonblocking
	}
}

// WithPanicHandler sets up panic handler.
func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(opts *Options) {
		opts.PanicHandler = panicHandler
	}
}

// WithLogger sets up a customized logger.
func WithLogger(logger Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

// WithDisablePurge indicates whether we turn off automatically purge.
func WithDisablePurge(disable bool) Option {
	return func(opts *Options) {
		opts.DisablePurge = disable
	}
}

// boost options ...

func WithSize(size uint) Option {
	return func(opts *Options) {
		opts.Size = size
	}
}

func WithGenerator(generator IDGenerator) Option {
	return func(opts *Options) {
		if generator != nil {
			opts.Generator = generator
		}
	}
}

func WithInput(size uint) Option {
	return func(opts *Options) {
		opts.Input.BufferSize = size
	}
}

func WithOutput(size uint, interval, timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Output = &OutputOptions{
			BufferSize:         size,
			CheckCloseInterval: max(interval, MinimumCheckCloseInterval),
			TimeoutOnSend:      max(timeout, MinimumTimeoutOnSend),
		}
	}
}
