package rx

import (
	"context"

	"github.com/teivah/onecontext"
)

var emptyContext context.Context

type Option[I any] interface {
	apply(*funcOption[I])
	// ...
	buildChannel() chan Item[I]
	buildContext(parent context.Context) context.Context
	// ...
	isConnectable() bool
	isConnectOperation() bool
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

func (fdo *funcOption[I]) isConnectable() bool {
	return fdo.connectable
}

func (fdo *funcOption[I]) isConnectOperation() bool {
	return fdo.connectOperation
}

func (fdo *funcOption[I]) apply(do *funcOption[I]) {
	fdo.f(do)
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
