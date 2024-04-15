package rx

import (
	"context"
	"reflect"
	"time"

	"github.com/snivilised/lorax/enums"
)

type (
	// Item is a wrapper having either a value, error or an auxiliary value.
	//
	Item[T any] struct {
		V    T
		E    error
		aux  any
		disc enums.ItemDiscriminator
	}

	// TimestampItem attach a timestamp to an item.
	//
	TimestampItem[T any] struct {
		Timestamp time.Time
		V         T
	}
)

// Of creates an item from a value.
func Of[T any](v T) Item[T] {
	return Item[T]{
		V:    v,
		disc: enums.ItemDiscNative,
	}
}

// Returns the native value of item
func (it Item[T]) Of() T {
	return it.aux.(T)
}

// Ch creates an item from a channel
func Ch[T any](ch any) Item[T] {
	if c, ok := ch.(chan<- Item[T]); ok {
		return Item[T]{
			aux:  c,
			disc: enums.ItemDiscChan,
		}
	}

	panic("temp: invalid ch type")
}

// Returns the channel value of item
func (it Item[T]) Ch() chan<- Item[T] {
	return it.aux.(chan<- Item[T])
}

// Error creates an item from an error.
func Error[T any](err error) Item[T] {
	return Item[T]{
		E:    err,
		disc: enums.ItemDiscError,
	}
}

// Returns the error value of item
func (it Item[T]) Error() error {
	return it.aux.(error)
}

// Returns the error value of item
func (it Item[T]) Pulse() error {
	return it.aux.(error)
}

// TV creates a type safe tick value instance
func TV[T any](tv int) Item[T] {
	return Item[T]{
		aux:  tv,
		disc: enums.ItemDiscTickValue,
	}
}

// Returns the tick value of item
func (it Item[T]) TV() int {
	return it.aux.(int)
}

// Num creates a type safe tick value instance
func Num[T any](n NumVal) Item[T] {
	return Item[T]{
		aux:  n,
		disc: enums.ItemDiscNumber,
	}
}

// Returns the tick value of item
func (it Item[T]) Num() NumVal {
	return it.aux.(NumVal)
}

// Bool creates a type safe boolean instance
func Bool[T any](b bool) Item[T] {
	return Item[T]{
		aux:  b,
		disc: enums.ItemDiscBoolean,
	}
}

// Returns the tick value of item
func (it Item[T]) Bool() bool {
	return it.aux.(bool)
}

// True creates a type safe boolean instance set to true
func True[T any]() Item[T] {
	return Item[T]{
		aux:  true,
		disc: enums.ItemDiscBoolean,
	}
}

// False creates a type safe boolean instance set to false
func False[T any]() Item[T] {
	return Item[T]{
		aux:  false,
		disc: enums.ItemDiscBoolean,
	}
}

// Opaque creates an item from any type of value
func Opaque[T any](o any) Item[T] {
	return Item[T]{
		aux:  o,
		disc: enums.ItemDiscOpaque,
	}
}

// Opaque returns the opaque value of item without typecast
func (it Item[T]) Opaque() any {
	return it.aux
}

// Disc returns the discriminator of the item
func (it Item[T]) Disc() enums.ItemDiscriminator {
	return it.disc
}

// SendItems is a utility function that sends a list of items and indicates a
// strategy on whether to close the channel once the function completes.
func SendItems[T any](ctx context.Context,
	ch chan<- Item[T], strategy enums.CloseChannelStrategy, items ...any,
) {
	if strategy == enums.CloseChannel {
		defer close(ch)
	}

	send(ctx, ch, items...)
}

func send[T any](ctx context.Context, ch chan<- Item[T], items ...any) {
	// can we revert items to be Item[T]?
	for _, current := range items {
		switch item := current.(type) {
		default:
			rt := reflect.TypeOf(item)
			sendItemByType(ctx, ch, rt, item)

		case error:
			Error[T](item).SendContext(ctx, ch)
		}
	}
}

func sendItemByType[T any](ctx context.Context, ch chan<- Item[T], rt reflect.Type, item any) {
	switch rt.Kind() { //nolint:exhaustive // foo
	default:
		switch v := item.(type) {
		case error:
			Error[T](v).SendContext(ctx, ch)

		case Item[T]:
			v.SendContext(ctx, ch)

		case T:
			Of(v).SendContext(ctx, ch)
		}

	case reflect.Chan:
		inCh := reflect.ValueOf(item) // current
		sendViaRefCh(ctx, inCh, ch)

	case reflect.Slice:
		s := reflect.ValueOf(item) // current

		for i := 0; i < s.Len(); i++ {
			send(ctx, ch, s.Index(i).Interface())
		}
	}
}

func sendViaRefCh[T any](ctx context.Context, inCh reflect.Value, ch chan<- Item[T]) {
	for {
		v, ok := inCh.Recv()

		if !ok {
			return
		}

		vItem := v.Interface()

		switch item := vItem.(type) {
		default:
			Ch[T](item).SendContext(ctx, ch)

		case error:
			Error[T](item).SendContext(ctx, ch)
		}
	}
}

// IsValue checks if an item is a value.
func (it Item[T]) IsValue() bool {
	return it.disc == 0
}

// IsCh checks if an item is an error.
func (it Item[T]) IsCh() bool {
	return (it.disc & enums.ItemDiscChan) > 0
}

// IsError checks if an item is an error.
func (it Item[T]) IsError() bool {
	return (it.disc & enums.ItemDiscError) > 0
}

// IsTick checks if an item is a tick instance.
func (it Item[T]) IsTick() bool {
	return (it.disc & enums.ItemDiscTick) > 0
}

// IsTickValue checks if an item is a tick value instance.
func (it Item[T]) IsTickValue() bool {
	return (it.disc & enums.ItemDiscTickValue) > 0
}

// IsNumber checks if an item is a numeric instance.
func (it Item[T]) IsNumber() bool {
	return (it.disc & enums.ItemDiscNumber) > 0
}

// IsBoolean checks if an item is a boolean instance.
func (it Item[T]) IsBoolean() bool {
	return (it.disc & enums.ItemDiscBoolean) > 0
}

// IsOpaque checks if an item is an opaque value.
func (it Item[T]) IsOpaque() bool {
	return (it.disc & enums.ItemDiscOpaque) > 0
}

func (it Item[T]) Desc() string {
	return enums.ItemDescriptions[it.disc]
}

// SendBlocking sends an item and blocks until it is sent.
func (it Item[T]) SendBlocking(ch chan<- Item[T]) {
	ch <- it
}

// SendContext sends an item and blocks until it is sent or a context canceled.
// It returns a boolean to indicate whether the item was sent.
func (it Item[T]) SendContext(ctx context.Context, ch chan<- Item[T]) bool {
	select {
	case <-ctx.Done(): // Context's done channel has the highest priority
		return false
	default:
		select {
		case <-ctx.Done():
			return false
		case ch <- it:
			return true
		}
	}
}

func (it Item[T]) SendOpContext(ctx context.Context, ch any) bool { // Item[operator[T]]
	_ = ctx
	_ = ch

	panic("SendOpContext: NOT-IMPL")
}

// SendNonBlocking sends an item without blocking.
// It returns a boolean to indicate whether the item was sent.
func (it Item[T]) SendNonBlocking(ch chan<- Item[T]) bool {
	select {
	default:
		return false
	case ch <- it:
		return true
	}
}
