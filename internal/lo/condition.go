package lo

// Ternary is a 1 line if/else statement.
// Play: https://go.dev/play/p/t-D7WBL44h2
func Ternary[T any](condition bool, ifOutput, elseOutput T) T {
	if condition {
		return ifOutput
	}

	return elseOutput
}

// TernaryF is a 1 line if/else statement whose options are functions
// Play: https://go.dev/play/p/AO4VW20JoqM
func TernaryF[T any](condition bool, ifFunc, elseFunc func() T) T {
	if condition {
		return ifFunc()
	}

	return elseFunc()
}

type IfElse[T any] struct {
	result T
	done   bool
}

// If
// Play: https://go.dev/play/p/WSw3ApMxhyW
func If[T any](condition bool, result T) *IfElse[T] {
	if condition {
		return &IfElse[T]{result, true}
	}

	var t T
	return &IfElse[T]{t, false}
}

// IfF
// Play: https://go.dev/play/p/WSw3ApMxhyW
func IfF[T any](condition bool, resultF func() T) *IfElse[T] {
	if condition {
		return &IfElse[T]{resultF(), true}
	}

	var t T
	return &IfElse[T]{t, false}
}

// ElseIf
// Play: https://go.dev/play/p/WSw3ApMxhyW
func (i *IfElse[T]) ElseIf(condition bool, result T) *IfElse[T] {
	if !i.done && condition {
		i.result = result
		i.done = true
	}

	return i
}

// ElseIfF
// Play: https://go.dev/play/p/WSw3ApMxhyW
func (i *IfElse[T]) ElseIfF(condition bool, resultF func() T) *IfElse[T] {
	if !i.done && condition {
		i.result = resultF()
		i.done = true
	}

	return i
}

// Else
// Play: https://go.dev/play/p/WSw3ApMxhyW
func (i *IfElse[T]) Else(result T) T {
	if i.done {
		return i.result
	}

	return result
}

// ElseF
// Play: https://go.dev/play/p/WSw3ApMxhyW
func (i *IfElse[T]) ElseF(resultF func() T) T {
	if i.done {
		return i.result
	}

	return resultF()
}

type SwitchCase[T comparable, R any] struct {
	predicate T
	result    R
	done      bool
}

// Switch is a pure functional switch/case/default statement.
// Play: https://go.dev/play/p/TGbKUMAeRUd
func Switch[T comparable, R any](predicate T) *SwitchCase[T, R] {
	var result R

	return &SwitchCase[T, R]{
		predicate,
		result,
		false,
	}
}

// Case
// Play: https://go.dev/play/p/TGbKUMAeRUd
func (s *SwitchCase[T, R]) Case(val T, result R) *SwitchCase[T, R] {
	if !s.done && s.predicate == val {
		s.result = result
		s.done = true
	}

	return s
}

// CaseF
// Play: https://go.dev/play/p/TGbKUMAeRUd
func (s *SwitchCase[T, R]) CaseF(val T, cb func() R) *SwitchCase[T, R] {
	if !s.done && s.predicate == val {
		s.result = cb()
		s.done = true
	}

	return s
}

// Default
// Play: https://go.dev/play/p/TGbKUMAeRUd
func (s *SwitchCase[T, R]) Default(result R) R {
	if !s.done {
		s.result = result
	}

	return s.result
}

// DefaultF
// Play: https://go.dev/play/p/TGbKUMAeRUd
func (s *SwitchCase[T, R]) DefaultF(cb func() R) R {
	if !s.done {
		s.result = cb()
	}

	return s.result
}
