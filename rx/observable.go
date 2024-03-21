package rx

import (
	"context"
)

type Observable[I any] interface {
	Iterable[I]
}

// ObservableImpl implements Observable.
type ObservableImpl[I any] struct {
	parent   context.Context
	iterable Iterable[I]
}
