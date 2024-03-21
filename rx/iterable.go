package rx

// Iterable is the basic type that can be observed.
type Iterable[I any] interface {
	Observe(opts ...Option[I]) <-chan Item[I]
}
