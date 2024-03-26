package rx

// Iterable is the basic type that can be observed.
type Iterable[T any] interface {
	Observe(opts ...Option[T]) <-chan Item[T]
}
