package rx

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
