package rx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cenkalti/backoff/v4"
)

func isZero[T any](limit T) bool {
	val := reflect.ValueOf(limit).Interface()
	zero := reflect.Zero(reflect.TypeOf(limit)).Interface()

	return val != zero
}

// All determines whether all items emitted by an Observable meet some criteria.
func (o *ObservableImpl[T]) All(predicate Predicate[T], opts ...Option[T]) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &allOperator[T]{
			predicate: predicate,
			all:       true,
		}
	}, forceSeq, bypassGather, opts...)
}

type allOperator[T any] struct {
	predicate Predicate[T]
	all       bool
}

func (op *allOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if !op.predicate(item) {
		False[T]().SendContext(ctx, dst)

		op.all = false

		operatorOptions.stop()
	}
}

func (op *allOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *allOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.all {
		True[T]().SendContext(ctx, dst)
	}
}

func (op *allOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if !item.IsBoolean() {
		// This panic is temporary
		panic(fmt.Sprintf("item: '%+v' is not a Boolean", item))
	}

	if !item.B {
		False[T]().SendContext(ctx, dst)

		op.all = false

		operatorOptions.stop()
	}
}

func (o *ObservableImpl[T]) Average(calc Calculator[T],
	opts ...Option[T],
) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &averageOperator[T]{
			calc: calc,
		}
	}, forceSeq, bypassGather, opts...)
}

type averageOperator[T any] struct {
	sum   T
	count T
	calc  Calculator[T]
}

func (op *averageOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if !item.IsValue() || item.IsError() {
		Error[T](IllegalInputError{
			error: fmt.Sprintf("expected item value, got: %v (%v)", item.Desc(), item)},
		).SendContext(ctx, dst)
	}

	op.sum = op.calc.Add(op.sum, item.V)
	op.count = op.calc.Inc(op.count)
}

func (op *averageOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *averageOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.calc.IsZero(op.count) {
		Of(op.calc.Zero()).SendContext(ctx, dst)
	} else {
		Of(op.calc.Div(op.sum, op.count)).SendContext(ctx, dst)
	}
}

func (op *averageOperator[T]) gatherNext(_ context.Context, item Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	_ = item

	// TODO(fix): v := item.V.(*averageOperator[T])
	// op.sum += v.sum
	// op.count += v.count
	//
	panic("averageOperator.gatherNext NOT-IMPL")
}

// BackOffRetry implements a backoff retry if a source Observable sends an error,
// resubscribe to it in the hopes that it will complete without error.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) BackOffRetry(backOffCfg backoff.BackOff, opts ...Option[T]) Observable[T] {
	option := parseOptions(opts...)
	next := option.buildChannel()
	ctx := option.buildContext(o.parent)
	f := func() error {
		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				close(next)

				return nil
			case i, ok := <-observe:
				if !ok {
					return nil
				}

				if i.IsError() {
					return i.E
				}

				i.SendContext(ctx, next)
			}
		}
	}

	go func() {
		if err := backoff.Retry(f, backOffCfg); err != nil {
			Error[T](err).SendContext(ctx, next)
			close(next)

			return
		}

		close(next)
	}()

	return &ObservableImpl[T]{
		iterable: newChannelIterable(next),
	}
}

// Connect instructs a connectable Observable to begin emitting items to its subscribers.
func (o *ObservableImpl[T]) Connect(ctx context.Context) (context.Context, Disposable) {
	ctx, cancel := context.WithCancel(ctx)
	o.Observe(WithContext[T](ctx), connect[T]())

	return ctx, Disposable(cancel)
}

// Run creates an Observer without consuming the emitted items.
func (o *ObservableImpl[T]) Run(opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go func() {
		defer close(dispose)

		observe := o.Observe(opts...)

		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-observe:
				if !ok {
					return
				}
			}
		}
	}()

	return dispose
}

func (o *ObservableImpl[T]) Contains(equal Predicate[T], opts ...Option[T]) Single[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &containsOperator[T]{
			equal:    equal,
			contains: false,
		}
	}, forceSeq, bypassGather, opts...)
}

type containsOperator[T any] struct {
	equal    Predicate[T]
	contains bool
}

func (op *containsOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.equal(item) {
		True[T]().SendContext(ctx, dst)

		op.contains = true

		operatorOptions.stop()
	}
}

func (op *containsOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *containsOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.contains {
		False[T]().SendContext(ctx, dst)
	}
}

func (op *containsOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if item.IsBoolean() && item.B {
		True[T]().SendContext(ctx, dst)
		operatorOptions.stop()

		op.contains = true
	}
}

func (o *ObservableImpl[T]) Count(opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &countOperator[T]{}
	}, forceSeq, bypassGather, opts...)
}

type countOperator[T any] struct {
	count int
}

func (op *countOperator[T]) next(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
	op.count++
}

func (op *countOperator[T]) err(_ context.Context, _ Item[T],
	_ chan<- Item[T], operatorOptions operatorOptions[T],
) {
	operatorOptions.stop()
}

func (op *countOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	Num[T](op.count).SendContext(ctx, dst)
}

func (op *countOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// DefaultIfEmpty returns an Observable that emits the items emitted by the source
// Observable or a specified default item if the source Observable is empty.
func (o *ObservableImpl[T]) DefaultIfEmpty(defaultValue T, opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &defaultIfEmptyOperator[T]{
			defaultValue: defaultValue,
			empty:        true,
		}
	}, forceSeq, bypassGather, opts...)
}

type defaultIfEmptyOperator[T any] struct {
	defaultValue T
	empty        bool
}

func (op *defaultIfEmptyOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	item.SendContext(ctx, dst)
}

func (op *defaultIfEmptyOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *defaultIfEmptyOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if op.empty {
		Of(op.defaultValue).SendContext(ctx, dst)
	}
}

func (op *defaultIfEmptyOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Distinct suppresses duplicate items in the original Observable and returns
// a new Observable.
func (o *ObservableImpl[T]) Distinct(apply Func[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &distinctOperator[T]{
			apply:  apply,
			keyset: make(map[interface{}]interface{}),
		}
	}, forceSeq, bypassGather, opts...)
}

type distinctOperator[T any] struct {
	apply  Func[T]
	keyset map[interface{}]interface{}
}

func (op *distinctOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	key, err := op.apply(ctx, item.V)
	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	_, ok := op.keyset[key]

	if !ok {
		item.SendContext(ctx, dst)
	}

	op.keyset[key] = nil
}

func (op *distinctOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *distinctOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *distinctOperator[T]) gatherNext(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if _, contains := op.keyset[item.V]; !contains {
		Of(item.V).SendContext(ctx, dst)

		op.keyset[item.V] = nil
	}
}

// DistinctUntilChanged suppresses consecutive duplicate items in the original Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) DistinctUntilChanged(apply Func[T], comparator Comparator[T],
	opts ...Option[T],
) Observable[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return observable(o.parent, o, func() operator[T] {
		return &distinctUntilChangedOperator[T]{
			apply:      apply,
			comparator: comparator,
		}
	}, forceSeq, bypassGather, opts...)
}

type distinctUntilChangedOperator[T any] struct {
	apply      Func[T]
	current    T
	comparator Comparator[T]
}

func (op *distinctUntilChangedOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T]) {
	key, err := op.apply(ctx, item.V)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	if op.comparator(op.current, key) != 0 {
		item.SendContext(ctx, dst)

		op.current = key
	}
}

func (op *distinctUntilChangedOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *distinctUntilChangedOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *distinctUntilChangedOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// DoOnCompleted registers a callback action that will be called once the
// Observable terminates.
func (o *ObservableImpl[T]) DoOnCompleted(completedFunc CompletedFunc,
	opts ...Option[T],
) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)
		defer completedFunc()

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-src:
				if !ok {
					return
				}

				if i.IsError() {
					return
				}
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// DoOnError registers a callback action that will be called if the
// Observable terminates abnormally.
func (o *ObservableImpl[T]) DoOnError(errFunc ErrFunc, opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-src:
				if !ok {
					return
				}

				if i.IsError() {
					errFunc(i.E)

					return
				}
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// DoOnNext registers a callback action that will be called on each item
// emitted by the Observable.
func (o *ObservableImpl[T]) DoOnNext(nextFunc NextFunc[T], opts ...Option[T]) Disposed {
	dispose := make(chan struct{})
	handler := func(ctx context.Context, src <-chan Item[T]) {
		defer close(dispose)

		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-src:
				if !ok {
					return
				}

				if i.IsError() {
					return
				}

				nextFunc(i.V)
			}
		}
	}

	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)

	go handler(ctx, o.Observe(opts...))

	return dispose
}

// ElementAt emits only item n emitted by an Observable.
// Cannot be run in parallel.
func (o *ObservableImpl[T]) ElementAt(index uint, opts ...Option[T]) Single[T] {
	const (
		forceSeq     = true
		bypassGather = false
	)

	return single(o.parent, o, func() operator[T] {
		return &elementAtOperator[T]{
			index: index,
		}
	}, forceSeq, bypassGather, opts...)
}

type elementAtOperator[T any] struct {
	index     uint
	takeCount int
	sent      bool
}

func (op *elementAtOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.takeCount == int(op.index) {
		item.SendContext(ctx, dst)

		op.sent = true

		operatorOptions.stop()

		return
	}

	op.takeCount++
}

func (op *elementAtOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *elementAtOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.sent {
		Error[T](&IllegalInputError{}).SendContext(ctx, dst)
	}
}

func (op *elementAtOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// Error returns the eventual Observable error.
// This method is blocking.
func (o *ObservableImpl[T]) Error(opts ...Option[T]) error {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	observe := o.iterable.Observe(opts...)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-observe:
			if !ok {
				return nil
			}

			if item.IsError() {
				return item.E
			}
		}
	}
}

// Errors returns an eventual list of Observable errors.
// This method is blocking
func (o *ObservableImpl[T]) Errors(opts ...Option[T]) []error {
	option := parseOptions(opts...)
	ctx := option.buildContext(o.parent)
	observe := o.iterable.Observe(opts...)
	errs := make([]error, 0)

	for {
		select {
		case <-ctx.Done():
			return []error{ctx.Err()}
		case item, ok := <-observe:
			if !ok {
				return errs
			}

			if item.IsError() {
				errs = append(errs, item.E)
			}
		}
	}
}

// Filter emits only those items from an Observable that pass a predicate test.
func (o *ObservableImpl[T]) Filter(apply Predicate[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return observable(o.parent, o, func() operator[T] {
		return &filterOperator[T]{apply: apply}
	}, forceSeq, bypassGather, opts...)
}

type filterOperator[T any] struct {
	apply Predicate[T]
}

func (op *filterOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], _ operatorOptions[T],
) {
	if op.apply(item) {
		item.SendContext(ctx, dst)
	}
}

func (op *filterOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *filterOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *filterOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T],
) {
}

// Find emits the first item passing a predicate then complete.
func (o *ObservableImpl[T]) Find(find Predicate[T], opts ...Option[T]) OptionalSingle[T] {
	const (
		forceSeq     = true
		bypassGather = true
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &findOperator[T]{
			find: find,
		}
	}, true, true, opts...)
}

type findOperator[T any] struct {
	find Predicate[T]
}

func (op *findOperator[T]) next(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	if op.find(item) {
		item.SendContext(ctx, dst)
		operatorOptions.stop()
	}
}

func (op *findOperator[T]) err(ctx context.Context, item Item[T],
	dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *findOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *findOperator[T]) gatherNext(_ context.Context, _ Item[T],
	_ chan<- Item[T], _ operatorOptions[T]) {
}

// !!!

// Max determines and emits the maximum-valued item emitted by an Observable according to a comparator.
func (o *ObservableImpl[T]) Max(comparator Comparator[T], initLimit InitLimit[T],
	opts ...Option[T],
) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &maxOperator[T]{
			comparator: comparator,
			max:        initLimit(),
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

type maxOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	max        T
}

func (op *maxOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	if op.comparator(op.max, item.V) < 0 {
		op.max = item.V
	}
}

func (op *maxOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *maxOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		// Using Num here instead of Of, means that the Max operator only works
		// for numbers and not any other Item type, which is probably not what
		// we wanted. We initially tried to limit the Max operator to number,
		// but in the fullness of time, this looks to be incorrect. (FuncN!!!)
		// If we had a channel of widgets, this wouldn't work when it supposed to be
		// able to, so we need a new design. But this was a valuable learning
		// experience. We need to radically redesign Min/Max/Map operators to
		// be able to work properly, and also can detect when a value has been
		// set. In the legacy rxgo, it depended on reflection and the ability
		// to perform a test like:
		//
		// if op.max == nil {
		// 	op.max = item.V
		// } else {
		// 	if op.comparator(op.max, item.V) < 0 {
		// 		op.max = item.V
		// 	}
		// }
		//
		// There should be no internal code that tries to compare values, because
		// when generics are in play, only the client knows how to do this, so
		// there should be a way for the client to implement these types of checks
		// themselves, probably by passing in a new function like comparator.
		//
		Of(op.max).SendContext(ctx, dst)
	}
}

func (op *maxOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*maxOperator).max), dst, operatorOptions)÷
	op.next(ctx, Of(item.V), dst, operatorOptions)
}

// Min determines and emits the minimum-valued item emitted by an Observable
// according to a comparator.
func (o *ObservableImpl[T]) Min(comparator Comparator[T], initLimit InitLimit[T],
	opts ...Option[T],
) OptionalSingle[T] {
	const (
		forceSeq     = false
		bypassGather = false
	)

	return optionalSingle(o.parent, o, func() operator[T] {
		return &minOperator[T]{
			min:        initLimit(),
			comparator: comparator,
			empty:      true,
		}
	}, forceSeq, bypassGather, opts...)
}

// Map transforms the items emitted by an Observable by applying a function to each item.
func (o *ObservableImpl[T]) Map(apply Func[T], opts ...Option[T]) Observable[T] {
	const (
		forceSeq     = false
		bypassGather = true
	)

	return observable(o.parent, o, func() operator[T] {
		return &mapOperator[T]{
			apply: apply,
		}
	}, forceSeq, bypassGather, opts...)
}

type mapOperator[T any] struct {
	apply Func[T]
}

func (op *mapOperator[T]) next(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	res, err := op.apply(ctx, item.V)

	if err != nil {
		Error[T](err).SendContext(ctx, dst)
		operatorOptions.stop()

		return
	}

	Of(res).SendContext(ctx, dst)
}

func (op *mapOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *mapOperator[T]) end(_ context.Context, _ chan<- Item[T]) {
}

func (op *mapOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], _ operatorOptions[T],
) {
	// switch item.V.(type) {
	// case *mapOperator:
	// 	return
	// }
	// TODO: check above switch not required
	item.SendContext(ctx, dst)
}

type minOperator[T any] struct {
	comparator Comparator[T]
	empty      bool
	min        T
	limit      func(value T) bool
}

func (op *minOperator[T]) next(_ context.Context,
	item Item[T], _ chan<- Item[T], _ operatorOptions[T],
) {
	op.empty = false

	if op.comparator(op.min, item.V) > 0 {
		op.min = item.V
	}
}

func (op *minOperator[T]) err(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	defaultErrorFuncOperator(ctx, item, dst, operatorOptions)
}

func (op *minOperator[T]) end(ctx context.Context, dst chan<- Item[T]) {
	if !op.empty {
		Of(op.min).SendContext(ctx, dst)
	}
}

func (op *minOperator[T]) gatherNext(ctx context.Context,
	item Item[T], dst chan<- Item[T], operatorOptions operatorOptions[T],
) {
	// TODO(check): op.next(ctx, Of(item.V.(*minOperator).min), dst, operatorOptions)
	op.next(ctx, Of(item.V), dst, operatorOptions)
}

func (o *ObservableImpl[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return o.iterable.Observe(opts...)
}
