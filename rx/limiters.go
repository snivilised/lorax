package rx

import (
	"math"

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

func MaxInitLimitInt() int {
	return math.MinInt
}

func MinInitLimitInt() int {
	return math.MaxInt
}

// TODO: add limiters for other Ordered constraint types
