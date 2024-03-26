package rx

import (
	"context"
	"sync"

	"golang.org/x/exp/constraints"
)

func LimitComparator[T constraints.Ordered](a, b T) int {
	if a == b {
		return 0
	}

	if a < b {
		return -1
	}

	return 1
}

type Observable[T any] interface {
	Iterable[T]

	Max(comparator Comparator[T], opts ...Option[T]) OptionalSingle[T]
	Min(comparator Comparator[T], opts ...Option[T]) OptionalSingle[T]
}

// ObservableImpl implements Observable.
type ObservableImpl[T any] struct {
	parent   context.Context
	iterable Iterable[T]
}

func defaultErrorFuncOperator[T any](ctx context.Context,
	item Item[T], dst chan<- Item[T], options operatorOptions[T],
) {
	item.SendContext(ctx, dst)
	options.stop()
}

type operator[T any] interface {
	next(ctx context.Context, item Item[T], dst chan<- Item[T], options operatorOptions[T])
	err(ctx context.Context, item Item[T], dst chan<- Item[T], options operatorOptions[T])
	end(ctx context.Context, dst chan<- Item[T])
	gatherNext(ctx context.Context, item Item[T], dst chan<- Item[T], options operatorOptions[T])
}

func single[T any](parent context.Context,
	iterable Iterable[T], operatorFactory func() operator[T], forceSeq, bypassGather bool, opts ...Option[T],
) Single[T] {
	option := parseOptions(opts...)
	parallel, _ := option.getPool()
	next := option.buildChannel()
	ctx := option.buildContext(parent)

	if option.isEagerObservation() {
		if forceSeq || !parallel {
			runSequential(ctx, next, iterable, operatorFactory, option, opts...)
		} else {
			runParallel(ctx, next, iterable.Observe(opts...), operatorFactory, bypassGather, option, opts...)
		}

		return &SingleImpl[T]{iterable: newChannelIterable(next)}
	}

	return &SingleImpl[T]{
		iterable: newFactoryIterable(func(propagatedOptions ...Option[T]) <-chan Item[T] {
			mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // foo

			option = parseOptions(mergedOptions...)

			if forceSeq || !parallel {
				runSequential(ctx, next, iterable, operatorFactory, option, mergedOptions...)
			} else {
				runParallel(ctx, next, iterable.Observe(mergedOptions...), operatorFactory, bypassGather, option, mergedOptions...)
			}
			return next
		}),
	}
}

func optionalSingle[T any](parent context.Context,
	iterable Iterable[T], operatorFactory func() operator[T],
	forceSeq, bypassGather bool,
	opts ...Option[T],
) OptionalSingle[T] {
	option := parseOptions(opts...)
	ctx := option.buildContext(parent)
	parallel, _ := option.getPool()

	if option.isEagerObservation() {
		next := option.buildChannel()

		if forceSeq || !parallel {
			runSequential(ctx, next, iterable,
				operatorFactory, option, opts...,
			)
		} else {
			runParallel(ctx, next, iterable.Observe(opts...),
				operatorFactory, bypassGather, option, opts...,
			)
		}

		return &OptionalSingleImpl[T]{iterable: newChannelIterable(next)}
	}

	return &OptionalSingleImpl[T]{
		parent: ctx,
		iterable: newFactoryIterable(func(propagatedOptions ...Option[T]) <-chan Item[T] {
			mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // foo
			option = parseOptions(mergedOptions...)

			next := option.buildChannel()
			ctx := option.buildContext(parent)

			if forceSeq || !parallel {
				runSequential(ctx, next, iterable,
					operatorFactory, option, mergedOptions...,
				)
			} else {
				runParallel(ctx, next, iterable.Observe(mergedOptions...),
					operatorFactory, bypassGather, option, mergedOptions...,
				)
			}

			return next
		}),
	}
}

func runSequential[T any](ctx context.Context,
	next chan Item[T], iterable Iterable[T], operatorFactory func() operator[T],
	option Option[T], opts ...Option[T],
) {
	observe := iterable.Observe(opts...)

	go func() {
		op := operatorFactory()
		stopped := false
		operator := operatorOptions[T]{
			stop: func() {
				if option.getErrorStrategy() == StopOnError {
					stopped = true
				}
			},
			resetIterable: func(newIterable Iterable[T]) {
				observe = newIterable.Observe(opts...)
			},
		}

	loop:
		for !stopped {
			select {
			case <-ctx.Done():
				break loop
			case i, ok := <-observe:
				if !ok {
					break loop
				}

				if i.IsError() {
					op.err(ctx, i, next, operator)
				} else {
					op.next(ctx, i, next, operator)
				}
			}
		}
		op.end(ctx, next)
		close(next)
	}()
}

func runParallel[T any](ctx context.Context,
	next chan Item[T], observe <-chan Item[T], operatorFactory func() operator[T],
	bypassGather bool, option Option[T], opts ...Option[T],
) {
	wg := sync.WaitGroup{}
	_, pool := option.getPool()
	wg.Add(pool)

	var gather chan Item[T]
	if bypassGather {
		gather = next
	} else {
		gather = make(chan Item[T], 1)

		// Gather
		go func() {
			op := operatorFactory()
			stopped := false
			operator := operatorOptions[T]{
				stop: func() {
					if option.getErrorStrategy() == StopOnError {
						stopped = true
					}
				},
				resetIterable: func(newIterable Iterable[T]) {
					observe = newIterable.Observe(opts...)
				},
			}

			for item := range gather {
				if stopped {
					break
				}

				if item.IsError() {
					op.err(ctx, item, next, operator)
				} else {
					op.gatherNext(ctx, item, next, operator)
				}
			}

			op.end(ctx, next)
			close(next)
		}()
	}

	// Scatter
	for i := 0; i < pool; i++ {
		go func() {
			op := operatorFactory()
			stopped := false
			operator := operatorOptions[T]{
				stop: func() {
					if option.getErrorStrategy() == StopOnError {
						stopped = true
					}
				},
				resetIterable: func(newIterable Iterable[T]) {
					observe = newIterable.Observe(opts...)
				},
			}

			defer wg.Done()

			for !stopped {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-observe:
					if !ok {
						if !bypassGather {
							// TODO:
							// cannot use gather (variable of type chan Item[T]) as chan<- Item[operator[T]]
							// value in argument to Of(op).SendContext
							//
							// op = operator[T] / Item[operator[T]]
							// gather = chan Item[T]
							// can we send T down the channel, then apply the operator on the other
							// end of the channel?
							//
							// or can we define another method on Item, such as SendOp ==> this looks
							// like a better option.
							// SendOpContext
							Of(op).SendOpContext(ctx, gather)
						}

						return
					}

					if item.IsError() {
						op.err(ctx, item, gather, operator)
					} else {
						op.next(ctx, item, gather, operator)
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(gather)
	}()
}
