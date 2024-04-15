package rx

import (
	"golang.org/x/exp/constraints"
)

// Numeric
type Numeric interface {
	constraints.Integer | constraints.Signed | constraints.Unsigned | constraints.Float
}

func Calc[T Numeric]() Calculator[T] {
	return new(NumericCalc[T])
}

// NumericCalc is a predefine calculator for any numeric type
type NumericCalc[T Numeric] struct {
	zero T
}

func (c *NumericCalc[T]) Add(a, b T) T {
	return a + b
}

func (c *NumericCalc[T]) Div(a, b T) T {
	if c.IsZero(b) {
		return c.Zero()
	}

	return a / b
}

func (c *NumericCalc[T]) Inc(v T) T {
	return v + T(1)
}

func (c *NumericCalc[T]) IsZero(v T) bool {
	return v == c.zero
}

func (c *NumericCalc[T]) Zero() T {
	return c.zero
}

func ItemCalc[T Numeric]() Calculator[T] {
	return new(NumericCalc[T])
}

type NumericItemCalc[T Numeric] struct {
	zero T
}

func (c *NumericItemCalc[T]) Add(a, b Item[T]) int {
	return a.Num() + b.Num()
}

func (c *NumericItemCalc[T]) Div(a, b Item[T]) Item[T] {
	if b.Num() == 0 {
		return c.Zero()
	}

	// !!!TODO(fix): we might need to use Opaque or another type
	// so we don't lose the precision with this division.
	//
	return Num[T](a.Num() / b.Num())
}

func (c *NumericItemCalc[T]) Inc(i Item[T]) Item[T] {
	n, _ := i.aux.(int)
	n++
	i.aux = n

	return i
}

func (c *NumericItemCalc[T]) IsZero(v Item[T]) bool {
	return v.Num() == 0
}

func (c *NumericItemCalc[T]) Zero() Item[T] {
	return Num[T](0)
}
