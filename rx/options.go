package rx

import (
	"context"
	"runtime"

	"github.com/teivah/onecontext"
)

var emptyContext context.Context

type Option[I any] interface {
	apply(*funcOption[I])
	toPropagate() bool
	isEagerObservation() bool
	getPool() (bool, int)
	buildChannel() chan Item[I]
	buildContext(parent context.Context) context.Context
	getBackPressureStrategy() BackPressureStrategy
	getErrorStrategy() OnErrorStrategy
	isConnectable() bool
	isConnectOperation() bool
	isSerialized() (bool, func(I) int)
}

type funcOption[I any] struct {
	f                    func(*funcOption[I])
	isBuffer             bool
	buffer               int
	ctx                  context.Context
	observation          ObservationStrategy
	pool                 int
	backPressureStrategy BackPressureStrategy
	onErrorStrategy      OnErrorStrategy
	propagate            bool
	connectable          bool
	connectOperation     bool
	serialized           func(I) int
}

func (fdo *funcOption[I]) toPropagate() bool {
	return fdo.propagate
}

func (fdo *funcOption[I]) isEagerObservation() bool {
	return fdo.observation == Eager
}

func (fdo *funcOption[I]) getPool() (b bool, p int) {
	return fdo.pool > 0, fdo.pool
}

func (fdo *funcOption[I]) buildChannel() chan Item[I] {
	if fdo.isBuffer {
		return make(chan Item[I], fdo.buffer)
	}

	return make(chan Item[I])
}

func (fdo *funcOption[I]) buildContext(parent context.Context) context.Context {
	if fdo.ctx != nil && parent != nil {
		ctx, _ := onecontext.Merge(fdo.ctx, parent)

		return ctx
	}

	if fdo.ctx != nil {
		return fdo.ctx
	}

	if parent != nil {
		return parent
	}

	return context.Background()
}

func (fdo *funcOption[I]) getBackPressureStrategy() BackPressureStrategy {
	return fdo.backPressureStrategy
}

func (fdo *funcOption[I]) getErrorStrategy() OnErrorStrategy {
	return fdo.onErrorStrategy
}

func (fdo *funcOption[I]) isConnectable() bool {
	return fdo.connectable
}

func (fdo *funcOption[I]) isConnectOperation() bool {
	return fdo.connectOperation
}

func (fdo *funcOption[I]) apply(do *funcOption[I]) {
	fdo.f(do)
}

func (fdo *funcOption[I]) isSerialized() (b bool, f func(I) int) {
	if fdo.serialized == nil {
		return false, nil
	}

	return true, fdo.serialized
}

func newFuncOption[I any](f func(*funcOption[I])) *funcOption[I] {
	return &funcOption[I]{
		f: f,
	}
}

func parseOptions[I any](opts ...Option[I]) Option[I] {
	o := new(funcOption[I])
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

// WithBufferedChannel allows to configure the capacity of a buffered channel.
func WithBufferedChannel[I any](capacity int) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.isBuffer = true
		options.buffer = capacity
	})
}

// WithContext allows to pass a context.
func WithContext[I any](ctx context.Context) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.ctx = ctx
	})
}

// WithObservationStrategy uses the eager observation mode meaning consuming the items even without subscription.
func WithObservationStrategy[I any](strategy ObservationStrategy) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.observation = strategy
	})
}

// WithPool allows to specify an execution pool.
func WithPool[I any](pool int) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.pool = pool
	})
}

// WithCPUPool allows to specify an execution pool based on the number of logical CPUs.
func WithCPUPool[I any]() Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.pool = runtime.NumCPU()
	})
}

// WithBackPressureStrategy sets the back pressure strategy: drop or block.
func WithBackPressureStrategy[I any](strategy BackPressureStrategy) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.backPressureStrategy = strategy
	})
}

// WithErrorStrategy defines how an observable should deal with error.
// This strategy is propagated to the parent observable.
func WithErrorStrategy[I any](strategy OnErrorStrategy) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.onErrorStrategy = strategy
	})
}

// WithPublishStrategy converts an ordinary Observable into a connectable Observable.
func WithPublishStrategy[I any]() Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.connectable = true
	})
}

// Serialize forces an Observable to make serialized calls and to be well-behaved.
func Serialize[I any](identifier func(I) int) Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.serialized = identifier
	})
}

func connect[I any]() Option[I] {
	return newFuncOption(func(options *funcOption[I]) {
		options.connectOperation = true
	})
}
