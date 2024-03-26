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

type Observable[I any] interface {
	Iterable[I]

	Max(comparator Comparator[I], opts ...Option[I]) OptionalSingle[I]
	Min(comparator Comparator[I], opts ...Option[I]) OptionalSingle[I]
}

// ObservableImpl implements Observable.
type ObservableImpl[I any] struct {
	parent   context.Context
	iterable Iterable[I]
}

func defaultErrorFuncOperator[I any](ctx context.Context,
	item Item[I], dst chan<- Item[I], options operatorOptions[I],
) {
	item.SendContext(ctx, dst)
	options.stop()
}

type operator[I any] interface {
	next(ctx context.Context, item Item[I], dst chan<- Item[I], options operatorOptions[I])
	err(ctx context.Context, item Item[I], dst chan<- Item[I], options operatorOptions[I])
	end(ctx context.Context, dst chan<- Item[I])
	gatherNext(ctx context.Context, item Item[I], dst chan<- Item[I], options operatorOptions[I])
}

func single[I any](parent context.Context,
	iterable Iterable[I], operatorFactory func() operator[I], forceSeq, bypassGather bool, opts ...Option[I],
) Single[I] {
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

		return &SingleImpl[I]{iterable: newChannelIterable(next)}
	}

	return &SingleImpl[I]{
		iterable: newFactoryIterable(func(propagatedOptions ...Option[I]) <-chan Item[I] {
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

func optionalSingle[I any](parent context.Context,
	iterable Iterable[I], operatorFactory func() operator[I],
	forceSeq, bypassGather bool,
	opts ...Option[I],
) OptionalSingle[I] {
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

		return &OptionalSingleImpl[I]{iterable: newChannelIterable(next)}
	}

	return &OptionalSingleImpl[I]{
		parent: ctx,
		iterable: newFactoryIterable(func(propagatedOptions ...Option[I]) <-chan Item[I] {
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

func runSequential[I any](ctx context.Context,
	next chan Item[I], iterable Iterable[I], operatorFactory func() operator[I],
	option Option[I], opts ...Option[I],
) {
	observe := iterable.Observe(opts...)

	go func() {
		op := operatorFactory()
		stopped := false
		operator := operatorOptions[I]{
			stop: func() {
				if option.getErrorStrategy() == StopOnError {
					stopped = true
				}
			},
			resetIterable: func(newIterable Iterable[I]) {
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

func runParallel[I any](ctx context.Context,
	next chan Item[I], observe <-chan Item[I], operatorFactory func() operator[I],
	bypassGather bool, option Option[I], opts ...Option[I],
) {
	wg := sync.WaitGroup{}
	_, pool := option.getPool()
	wg.Add(pool)

	var gather chan Item[I]
	if bypassGather {
		gather = next
	} else {
		gather = make(chan Item[I], 1)

		// Gather
		go func() {
			op := operatorFactory()
			stopped := false
			operator := operatorOptions[I]{
				stop: func() {
					if option.getErrorStrategy() == StopOnError {
						stopped = true
					}
				},
				resetIterable: func(newIterable Iterable[I]) {
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
			operator := operatorOptions[I]{
				stop: func() {
					if option.getErrorStrategy() == StopOnError {
						stopped = true
					}
				},
				resetIterable: func(newIterable Iterable[I]) {
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
							// cannot use gather (variable of type chan Item[I]) as chan<- Item[operator[I]]
							// value in argument to Of(op).SendContext
							//
							// op = operator[I] / Item[operator[I]]
							// gather = chan Item[I]
							// can we send I down the channel, then apply the operator on the other
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
