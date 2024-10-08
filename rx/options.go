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
	"runtime"

	"github.com/snivilised/lorax/enums"
	"github.com/teivah/onecontext"
)

var emptyContext context.Context

type Option[T any] interface {
	apply(*funcOption[T])
	toPropagate() bool
	isEagerObservation() bool
	getPool() (bool, int)
	buildChannel() chan Item[T]
	buildContext(parent context.Context) context.Context
	getBackPressureStrategy() enums.BackPressureStrategy
	getErrorStrategy() enums.OnErrorStrategy
	isConnectable() bool
	isConnectOperation() bool
	isSerialized() (bool, func(T) int)
	calc() Calculator[T]
}

type funcOption[T any] struct {
	f                    func(*funcOption[T])
	isBuffer             bool
	buffer               int
	ctx                  context.Context
	observation          enums.ObservationStrategy
	pool                 int
	backPressureStrategy enums.BackPressureStrategy
	onErrorStrategy      enums.OnErrorStrategy
	propagate            bool
	connectable          bool
	connectOperation     bool
	serialized           func(T) int
	calculator           Calculator[T]
}

func (fdo *funcOption[T]) toPropagate() bool {
	return fdo.propagate
}

func (fdo *funcOption[T]) isEagerObservation() bool {
	return fdo.observation == enums.Eager
}

func (fdo *funcOption[T]) getPool() (b bool, p int) {
	return fdo.pool > 0, fdo.pool
}

func (fdo *funcOption[T]) buildChannel() chan Item[T] {
	if fdo.isBuffer {
		return make(chan Item[T], fdo.buffer)
	}

	return make(chan Item[T])
}

func (fdo *funcOption[T]) buildContext(parent context.Context) context.Context {
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

func (fdo *funcOption[T]) getBackPressureStrategy() enums.BackPressureStrategy {
	return fdo.backPressureStrategy
}

func (fdo *funcOption[T]) getErrorStrategy() enums.OnErrorStrategy {
	return fdo.onErrorStrategy
}

func (fdo *funcOption[T]) isConnectable() bool {
	return fdo.connectable
}

func (fdo *funcOption[T]) isConnectOperation() bool {
	return fdo.connectOperation
}

func (fdo *funcOption[T]) apply(do *funcOption[T]) {
	fdo.f(do)
}

func (fdo *funcOption[T]) isSerialized() (b bool, f func(T) int) {
	if fdo.serialized == nil {
		return false, nil
	}

	return true, fdo.serialized
}

func (fdo *funcOption[T]) calc() Calculator[T] {
	return fdo.calculator
}

func newFuncOption[T any](f func(*funcOption[T])) *funcOption[T] {
	return &funcOption[T]{
		f: f,
	}
}

func parseOptions[T any](opts ...Option[T]) Option[T] {
	o := new(funcOption[T])
	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

// WithBufferedChannel allows to configure the capacity of a buffered channel.
func WithBufferedChannel[T any](capacity int) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.isBuffer = true
		options.buffer = capacity
	})
}

// WithContext allows to pass a context.
func WithContext[T any](ctx context.Context) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.ctx = ctx
	})
}

// WithObservationStrategy uses the eager observation mode meaning consuming the items even without subscription.
func WithObservationStrategy[T any](strategy enums.ObservationStrategy) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.observation = strategy
	})
}

// WithPool allows to specify an execution pool.
func WithPool[T any](pool int) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.pool = pool
	})
}

// WithCPUPool allows to specify an execution pool based on the number of logical CPUs.
func WithCPUPool[T any]() Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.pool = runtime.NumCPU()
	})
}

// WithBackPressureStrategy sets the back pressure strategy: drop or block.
func WithBackPressureStrategy[T any](strategy enums.BackPressureStrategy) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.backPressureStrategy = strategy
	})
}

// WithErrorStrategy defines how an observable should deal with error.
// This strategy is propagated to the parent observable.
func WithErrorStrategy[T any](strategy enums.OnErrorStrategy) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.onErrorStrategy = strategy
	})
}

// WithPublishStrategy converts an ordinary Observable into a connectable Observable.
func WithPublishStrategy[T any]() Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.connectable = true
	})
}

// Serialize forces an Observable to make serialized calls and to be well-behaved.
func Serialize[T any](identifier func(T) int) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.serialized = identifier
	})
}

func WithCalc[T any](calculator Calculator[T]) Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.calculator = calculator
	})
}

func connect[T any]() Option[T] {
	return newFuncOption(func(options *funcOption[T]) {
		options.connectOperation = true
	})
}
