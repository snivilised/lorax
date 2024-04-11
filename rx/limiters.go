package rx

import (
	"math"

	"golang.org/x/exp/constraints"
)

func NativeItemLimitComparator[T constraints.Ordered](a, b Item[T]) int {
	if a.V == b.V {
		return 0
	}

	if a.V < b.V {
		return -1
	}

	return 1
}

func NumericItemLimitComparator[T constraints.Ordered](a, b Item[T]) int {
	if a.N == b.N {
		return 0
	}

	if a.N < b.N {
		return -1
	}

	return 1
}

func MaxInitLimitInt() int {
	return math.MinInt
}

func MinInitLimitInt() int {
	return math.MaxInt
}

func MaxItemInitLimitInt() Item[int] {
	return Of(math.MinInt)
}

func MinItemInitLimitInt() Item[int] {
	return Of(math.MaxInt)
}

func MaxNItemInitLimitInt() Item[int] {
	return Num[int](math.MinInt)
}

func MinNItemInitLimitInt() Item[int] {
	return Num[int](math.MaxInt)
}

// TODO: add limiters for other Ordered constraint types
