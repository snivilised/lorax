package rx

// MIT License

// Copyright (c) 2016 Joe Chasinga

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"context"
	"reflect"
	"time"

	"github.com/snivilised/lorax/enums"
)

type (
	// Item is a wrapper having either a value, error or an opaque value
	//
	Item[T any] struct {
		V      T
		E      error
		opaque any
		disc   enums.ItemDiscriminator
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

// Zero creates a zero value item.
func Zero[T any]() Item[T] {
	var (
		zero T
	)

	return Item[T]{
		V:    zero,
		disc: enums.ItemDiscNative,
	}
}

// Error creates an item from an error.
func Error[T any](err error) Item[T] {
	return Item[T]{
		E:    err,
		disc: enums.ItemDiscError,
	}
}

// Bool creates a type safe boolean instance
func Bool[T any](b bool) Item[T] {
	return Item[T]{
		opaque: b,
		disc:   enums.ItemDiscBoolean,
	}
}

// Returns the tick value of item
func (it Item[T]) Bool() bool {
	return it.opaque.(bool)
}

// True creates a type safe boolean instance set to true
func True[T any]() Item[T] {
	return Item[T]{
		opaque: true,
		disc:   enums.ItemDiscBoolean,
	}
}

// False creates a type safe boolean instance set to false
func False[T any]() Item[T] {
	return Item[T]{
		opaque: false,
		disc:   enums.ItemDiscBoolean,
	}
}

// WCh creates an item from a channel
func WCh[T any](ch any) Item[T] {
	if c, ok := ch.(chan<- Item[T]); ok {
		return Item[T]{
			opaque: c,
			disc:   enums.ItemDiscWChan,
		}
	}

	panic("temp: invalid ch type")
}

// Returns the channel value of item
func (it Item[T]) Ch() chan<- Item[T] {
	return it.opaque.(chan<- Item[T])
}

// Num creates a type safe tick value instance
func Num[T any](n NumVal) Item[T] {
	return Item[T]{
		opaque: n,
		disc:   enums.ItemDiscNumber,
	}
}

// Returns the tick value of item
func (it Item[T]) Num() NumVal {
	return it.opaque.(NumVal)
}

// Opaque creates an item from any type of value
func Opaque[T any](o any) Item[T] {
	return Item[T]{
		opaque: o,
		disc:   enums.ItemDiscOpaque,
	}
}

// Opaque returns the opaque value of item without typecast
func (it Item[T]) Opaque() any {
	return it.opaque
}

// TV creates a type safe tick instance (no value)
func Tick[T any]() Item[T] {
	return Item[T]{
		disc: enums.ItemDiscTick,
	}
}

// TV creates a type safe tick value instance
func TV[T any](tv int) Item[T] {
	return Item[T]{
		opaque: tv,
		disc:   enums.ItemDiscTickValue,
	}
}

// Returns the tick value of item
func (it Item[T]) TV() int {
	return it.opaque.(int)
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
			WCh[T](item).SendContext(ctx, ch)

		case error:
			Error[T](item).SendContext(ctx, ch)
		}
	}
}

// IsValue checks if an item is a value.
func (it Item[T]) IsValue() bool {
	return (it.disc & enums.ItemDiscNative) > 0
}

// IsCh checks if an item is an error.
func (it Item[T]) IsCh() bool {
	return (it.disc & enums.ItemDiscWChan) > 0
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
