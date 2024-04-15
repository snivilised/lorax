package rx

import (
	"github.com/snivilised/lorax/enums"
)

type tryFunc[T, O any] func(Item[T]) (O, error)

func must[T, O any](item Item[T], fn tryFunc[T, O]) (result O) {
	var (
		err error
	)

	if result, err = fn(item); err != nil {
		panic(err)
	}

	return result
}

func try[T, O any](item Item[T], expected enums.ItemDiscriminator) (O, error) {
	if value, ok := item.aux.(O); ok {
		return value, nil
	}

	var (
		zero O
	)

	return zero, TryError[T]{
		item:     item,
		expected: expected,
	}
}

// TryBool is a helper function that assists the client in interpreting
// a boolean item received via a channel. If the item it not a boolean,
// an error is returned, otherwise the returned value is the underlying
// boolean value.
func TryBool[T any](item Item[T]) (bool, error) {
	return try[T, bool](item, enums.ItemDiscBoolean)
}

// MustBool is a helper function that assists the client in interpreting
// a boolean item received via a channel. If the item it not a boolean,
// a panic is raised, otherwise the returned value is the underlying
// boolean value.
func MustBool[T any](item Item[T]) bool {
	return must(item, TryBool[T])
}

// TryWCh is a helper function that assists the client in interpreting
// a write chan item received via a channel. If the item it not a write chan,
// an error is returned, otherwise the returned value is the underlying
// write chan value.
func TryWCh[T any](item Item[T]) (chan<- Item[T], error) {
	return try[T, chan<- Item[T]](item, enums.ItemDiscWChan)
}

// MustWCh is a helper function that assists the client in interpreting
// a write chan item received via a channel. If the item it not a write chan,
// a panic is raised, otherwise the returned value is the underlying
// write chan value.
func MustWCh[T any](item Item[T]) chan<- Item[T] {
	return must(item, TryWCh[T])
}

// TryNum is a helper function that assists the client in interpreting
// an integer item received via a channel. If the item it not an integer,
// an error is returned, otherwise the returned value is the underlying
// integer value.
func TryNum[T any](item Item[T]) (NumVal, error) {
	return try[T, NumVal](item, enums.ItemDiscNumber)
}

// MustNum is a helper function that assists the client in interpreting
// an integer item received via a channel. If the item it not an integer,
// a panic is raised, otherwise the returned value is the underlying
// integer value.
func MustNum[T any](item Item[T]) NumVal {
	return must(item, TryNum[T])
}

// TryOpaque is a helper function that assists the client in interpreting
// an opaque item received via a channel. If the item it not of the custom
// type O, an error is returned, otherwise the returned value is the underlying
// custom value.
func TryOpaque[T, O any](item Item[T]) (O, error) {
	return try[T, O](item, enums.ItemDiscOpaque)
}

// MustOpaque is a helper function that assists the client in interpreting
// an opaque item received via a channel. If the item it not of the custom
// type O, a panic is raised, otherwise the returned value is the underlying
// custom value.
func MustOpaque[T, O any](item Item[T]) O {
	return must[T, O](item, TryOpaque[T])
}

// TryTV is a helper function that assists the client in interpreting
// an integer based tick value item received via a channel. If the item
// it not a tick value, an error is returned, otherwise the returned value
// is the underlying integer value.
func TryTV[T any](item Item[T]) (int, error) {
	return try[T, int](item, enums.ItemDiscTickValue)
}

// TryTV is a helper function that assists the client in interpreting
// an integer based tick value item received via a channel. If the item
// it not a tick value, a panic is raised, otherwise the returned value
// is the underlying integer value.
func MustTV[T any](item Item[T]) int {
	return must(item, TryTV[T])
}
