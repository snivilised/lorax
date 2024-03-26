package rx

type factoryIterable[T any] struct {
	factory func(opts ...Option[T]) <-chan Item[T]
}

func newFactoryIterable[T any](factory func(opts ...Option[T]) <-chan Item[T]) Iterable[T] {
	return &factoryIterable[T]{factory: factory}
}

func (i *factoryIterable[T]) Observe(opts ...Option[T]) <-chan Item[T] {
	return i.factory(opts...)
}
