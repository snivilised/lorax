package rx

func (o *ObservableImpl[I]) Observe(opts ...Option[I]) <-chan Item[I] {
	return o.iterable.Observe(opts...)
}
