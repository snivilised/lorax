package rx

type factoryIterable[I any] struct {
	factory func(opts ...Option[I]) <-chan Item[I]
}

func newFactoryIterable[I any](factory func(opts ...Option[I]) <-chan Item[I]) Iterable[I] {
	return &factoryIterable[I]{factory: factory}
}

func (i *factoryIterable[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	return i.factory(opts...)
}
