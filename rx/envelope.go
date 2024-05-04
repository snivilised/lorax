package rx

// Envelope wraps a struct type T so that it can be defined with
// pointer receivers. It also means that the client does not need
// to manually defined methods for constraint ProxyField as they
// are implemented by the Envelope.
type Envelope[T any, O Numeric] struct {
	T *T
	P O
}

// Seal wraps a struct T inside an Envelope
func Seal[T any, O Numeric](t *T) *Envelope[T, O] {
	// When using rx, we must instantiate it with a struct, not a pointer
	// to struct. So we create a distinction between what is the unit of
	// transfer (Envelope) and it's payload of type T. This means that the
	// Envelope can only have non pointer receivers, but T can
	// be defined with pointer receivers.
	//
	return &Envelope[T, O]{
		T: t,
	}
}

// Field nominates which member of T of type O is the proxy field required
// to satisfy constraint ProxyField.
func (e Envelope[T, O]) Field() O {
	return e.P
}

// Inc increments the P member of the Envelope required to satisfy constraint ProxyField.
func (e Envelope[T, O]) Inc(index *Envelope[T, O], by Envelope[T, O]) *Envelope[T, O] {
	index.P += by.P

	return index
}

// Index creates a new Envelope from the numeric value of i, required to
// satisfy constraint ProxyField.
func (e Envelope[T, O]) Index(i int) *Envelope[T, O] {
	return &Envelope[T, O]{
		P: O(i),
	}
}
