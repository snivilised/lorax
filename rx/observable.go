package rx

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cenkalti/backoff/v4"
	"github.com/emirpasic/gods/trees/binaryheap"
)

type Observable[T any] interface {
	Iterable[T]
	All(predicate Predicate[T], opts ...Option[T]) Single[T]
	Average(calc Calculator[T], opts ...Option[T]) Single[T]
	BackOffRetry(backOffCfg backoff.BackOff, opts ...Option[T]) Observable[T]
	Connect(ctx context.Context) (context.Context, Disposable)
	Contains(equal Predicate[T], opts ...Option[T]) Single[T]
	Count(opts ...Option[T]) Single[T]
	DefaultIfEmpty(defaultValue T, opts ...Option[T]) Observable[T]
	Distinct(apply Func[T], opts ...Option[T]) Observable[T]
	DistinctUntilChanged(apply Func[T], comparator Comparator[T], opts ...Option[T]) Observable[T]
	DoOnCompleted(completedFunc CompletedFunc, opts ...Option[T]) Disposed
	DoOnError(errFunc ErrFunc, opts ...Option[T]) Disposed
	DoOnNext(nextFunc NextFunc[T], opts ...Option[T]) Disposed
	Max(comparator Comparator[T], initLimit InitLimit[T], opts ...Option[T]) OptionalSingle[T]
	Map(apply Func[T], opts ...Option[T]) Observable[T]
	Min(comparator Comparator[T], initLimit InitLimit[T], opts ...Option[T]) OptionalSingle[T]

	Run(opts ...Option[T]) Disposed
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

func observable[T any](parent context.Context,
	iterable Iterable[T], operatorFactory func() operator[T], forceSeq, bypassGather bool, opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	parallel, _ := option.getPool()

	if option.isEagerObservation() {
		next := option.buildChannel()
		ctx := option.buildContext(parent)

		if forceSeq || !parallel {
			runSequential(ctx, next, iterable, operatorFactory, option, opts...)
		} else {
			runParallel(ctx, next, iterable.Observe(opts...), operatorFactory, bypassGather, option, opts...)
		}

		return &ObservableImpl[T]{iterable: newChannelIterable(next)}
	}

	if forceSeq || !parallel {
		return &ObservableImpl[T]{
			iterable: newFactoryIterable(func(propagatedOptions ...Option[T]) <-chan Item[T] {
				mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // ignore
				option := parseOptions(mergedOptions...)            //nolint:govet // shadow is deliberate
				next := option.buildChannel()
				ctx := option.buildContext(parent)

				runSequential(ctx, next, iterable, operatorFactory, option, mergedOptions...)

				return next
			}),
		}
	}

	if serialized, f := option.isSerialized(); serialized {
		firstItemIDCh := make(chan Item[T], 1)
		fromCh := make(chan Item[T], 1)
		obs := &ObservableImpl[T]{
			iterable: newFactoryIterable(func(propagatedOptions ...Option[T]) <-chan Item[T] {
				mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // ignore
				option := parseOptions(mergedOptions...)
				next := option.buildChannel()
				ctx := option.buildContext(parent)
				observe := iterable.Observe(opts...)

				go func() {
					select {
					case <-ctx.Done():
						return
					case firstItemID := <-firstItemIDCh:
						if firstItemID.IsError() {
							firstItemID.SendContext(ctx, fromCh)
							return
						}
						Of(firstItemID.V).SendContext(ctx, fromCh)
						// TODO: check int def here: Of(firstItemID.V.(int)).SendContext(ctx, fromCh)
						runParallel(ctx, next, observe, operatorFactory, bypassGather, option, mergedOptions...)
					}
				}()
				runFirstItem(ctx, f, firstItemIDCh, observe, next, operatorFactory, option, mergedOptions...)

				return next
			}),
		}

		return obs.serialize(parent, fromCh, f)
	}

	return &ObservableImpl[T]{
		iterable: newFactoryIterable(func(propagatedOptions ...Option[T]) <-chan Item[T] {
			mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // ignore
			option := parseOptions(mergedOptions...)
			next := option.buildChannel()
			ctx := option.buildContext(parent)

			runParallel(ctx, next, iterable.Observe(mergedOptions...),
				operatorFactory, bypassGather, option, mergedOptions...,
			)

			return next
		}),
	}
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
			mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // ignore
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
			mergedOptions := append(opts, propagatedOptions...) //nolint:gocritic // ignore
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

func runFirstItem[T any](ctx context.Context,
	f func(T) int, // TODO(check, return type int): func(T) int
	notif chan Item[T], observe <-chan Item[T], next chan Item[T],
	operatorFactory func() operator[T], option Option[T], opts ...Option[T],
) {
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
					i.SendContext(ctx, notif)
				} else {
					op.next(ctx, i, next, operator)
					Num[T](f(i.V)).SendContext(ctx, notif)
					// TODO(check this correct): Of[T](f(i.V)).SendContext(ctx, notif)
				}
			}
		}
		op.end(ctx, next)
	}()
}

func (o *ObservableImpl[T]) serialize(parent context.Context,
	fromCh chan Item[T], identifier func(T) int, opts ...Option[T],
) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()

	ctx := option.buildContext(parent)
	minHeap := binaryheap.NewWith(func(a, b interface{}) int {
		return a.(int) - b.(int)
	})
	items := make(map[int]interface{}) // TODO(check interface{} is correct, T?)

	var (
		from    int
		counter int64
	)

	src := o.Observe(opts...)

	go func() {
		select {
		case <-ctx.Done():
			close(next)

			return
		case item := <-fromCh:
			if item.IsError() {
				item.SendContext(ctx, next)
				close(next)

				return
			}

			from = item.N
			counter = int64(from)

			go func() {
				defer close(next)

				for {
					select {
					case <-ctx.Done():
						return
					case item, ok := <-src:
						if !ok {
							return
						}

						if item.IsError() {
							next <- item

							return
						}

						id := identifier(item.V)
						minHeap.Push(id)

						items[id] = item.V

						for !minHeap.Empty() {
							v, _ := minHeap.Peek()
							id, _ := v.(int)

							if atomic.LoadInt64(&counter) == int64(id) {
								if itemValue, contains := items[id]; contains {
									minHeap.Pop()
									delete(items, id)
									Num[T](itemValue.(int)).SendContext(ctx, next) // TODO(check me)

									counter++

									continue
								}
							}

							break
						}
					}
				}
			}()
		}
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}
