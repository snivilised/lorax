package boost

type observable[O any] struct {
	stream JobStreamR[O]
}

func (o *observable[O]) Observe() {
}
