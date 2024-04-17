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

// type rangeIterableL[T any] struct {
// 	start, count NumVal
// 	opts         []Option[T]
// }

type rangeIterable[T any] struct {
	start, count NumVal
	iterator     RangeIterator[T]
	opts         []Option[T]
}

func newRangeIterable[T any](iterator RangeIterator[T], opts ...Option[T]) Iterable[T] {
	return &rangeIterable[T]{
		iterator: iterator,
		opts:     opts,
	}
}

func (i *rangeIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()

	go func() {
		for idx, _ := i.iterator.Start(); i.iterator.While(*idx); i.iterator.Increment(idx) {
			select {
			case <-ctx.Done():
				return
			case next <- Of(*idx):
			}
		}
		close(next)
	}()

	return next
}

type rangeIterableNF[T NominatedField[T, O], O Numeric] struct {
	iterator RangeIteratorNF[T, O]
	opts     []Option[T]
}

func newRangeIterableNF[T NominatedField[T, O], O Numeric](iterator RangeIteratorNF[T, O],
	opts ...Option[T],
) Iterable[T] {
	return &rangeIterableNF[T, O]{
		iterator: iterator,
		opts:     opts,
	}
}

func (i *rangeIterableNF[T, O]) Observe(opts ...Option[T]) <-chan Item[T] {
	option := parseOptions(append(i.opts, opts...)...)
	ctx := option.buildContext(emptyContext)
	next := option.buildChannel()

	go func() {
		for idx, _ := i.iterator.Start(); i.iterator.While(*idx); i.iterator.Increment(idx) {
			select {
			case <-ctx.Done():
				return
			case next <- Of(*idx):
			}
		}

		close(next)
	}()

	return next
}

func LessThan[T Numeric](until T) WhilstFunc[T] {
	return func(current T) bool {
		return current < until
	}
}

func MoreThan[T Numeric](until T) WhilstFunc[T] {
	return func(current T) bool {
		return current > until
	}
}

func Count[T Numeric](count T) WhilstFunc[T] {
	return func(current T) bool {
		return current < count
	}
}

type NumericRangeIterator[T Numeric] struct {
	StartAt T
	StepBy  T
	Whilst  WhilstFunc[T]
	zero    T
}

func (i *NumericRangeIterator[T]) Init() error {
	if i.Whilst == nil {
		return RangeMissingWhilstError
	}

	return nil
}

// Start should return the initial index value. If the StepBy has
// not been set, it will default to 1.
func (i *NumericRangeIterator[T]) Start() (*T, error) {
	if i.StepBy == 0 {
		i.StepBy = 1
	}

	if i.Whilst == nil {
		return &i.zero, BadRangeIteratorError{}
	}

	return &i.StartAt, nil
}

func (i *NumericRangeIterator[T]) Step() T {
	return i.StepBy
}

// Increment increments the index value
func (i *NumericRangeIterator[T]) Increment(index *T) T {
	*(index) += i.StepBy

	return *(index)
}

// While defines a condition that must be true for the loop to
// continue iterating.
func (i *NumericRangeIterator[T]) While(current T) bool {
	return i.Whilst(current)
}

type NominatedRangeIterator[T NominatedField[T, O], O Numeric] struct {
	StartAt T
	StepBy  T
	Whilst  WhilstFunc[T]
	zero    T
}

func (i *NominatedRangeIterator[T, O]) Init() error {
	if i.Whilst == nil {
		return RangeMissingWhilstError
	}

	return nil
}

// Start should return the initial index value. If the StepBy has
// not been set, a panic occurs
func (i *NominatedRangeIterator[T, O]) Start() (*T, error) {
	if i.StepBy.Field() == 0 {
		panic("bad step-by, can't be zero")
	}

	if i.Whilst == nil {
		return &i.zero, BadRangeIteratorError{}
	}

	index := i.StartAt

	return &index, nil
}

func (i *NominatedRangeIterator[T, O]) Step() O {
	return i.StepBy.Field()
}

// Increment increments index value
func (i *NominatedRangeIterator[T, O]) Increment(index *T) *T {
	// This does look a bit strange but its a work around
	// for the fact that the instance of T is implemented with
	// non-pointer receivers and therefore can't make modifications
	// to itself (increment the index). We can't allow T to have
	// pointer receivers because that would be inappropriate for
	// scalar types (plus other issues), hence we are left with
	// this messy work-around, going via the back-door.
	//
	// index receives a pointer to a copy of itself, via Increment
	// and increments the copy.
	//
	(*index).Inc(index, i.StepBy)

	return index
}

// While defines a condition that must be true for the loop to
// continue iterating.
func (i *NominatedRangeIterator[T, O]) While(current T) bool {
	return i.Whilst(current)
}

func LessThanNF[T NominatedField[T, O], O Numeric](until T) WhilstFunc[T] {
	return func(current T) bool {
		return current.Field() < until.Field()
	}
}

func MoreThanNF[T NominatedField[T, O], O Numeric](until T) WhilstFunc[T] {
	return func(current T) bool {
		return current.Field() > until.Field()
	}
}
