package rx

import (
	"context"
	"reflect"
	"time"

	"github.com/snivilised/lorax/enums"
)

type (
	// Item is a wrapper having either a value, error or channel.
	//
	Item[T any] struct {
		Disc enums.ItemDiscriminator
		V    T
		E    error
		//
		C chan<- Item[T]
		N int
	}

	// TimestampItem attach a timestamp to an item.
	//
	TimestampItem[T any] struct {
		Timestamp time.Time
		V         T
	}

	// CloseChannelStrategy indicates a strategy on whether to close a channel.
	CloseChannelStrategy uint32
)

const (
	// LeaveChannelOpen indicates to leave the channel open after completion.
	LeaveChannelOpen CloseChannelStrategy = iota
	// CloseChannel indicates to close the channel open after completion.
	CloseChannel
)

// Of creates an item from a value.
func Of[T any](v T) Item[T] {
	return Item[T]{
		V:    v,
		Disc: enums.ItemDiscNative,
	}
}

// Ch creates an item from a channel
func Ch[T any](ch any) Item[T] {
	if c, ok := ch.(chan<- Item[T]); ok {
		return Item[T]{
			C:    c,
			Disc: enums.ItemDiscChan,
		}
	}

	panic("temp: invalid ch type")
}

// Error creates an item from an error.
func Error[T any](err error) Item[T] {
	return Item[T]{
		E:    err,
		Disc: enums.ItemDiscError,
	}
}

// Pulse creates a type safe tick instance that doesn't contain a value
// thats acts like a heartbeat.
func Pulse[T any]() Item[T] {
	return Item[T]{
		Disc: enums.ItemDiscPulse,
	}
}

// TV creates a type safe tick value instance
func TV[T any](tv int) Item[T] {
	return Item[T]{
		N:    tv,
		Disc: enums.ItemDiscTickValue,
	}
}

// Num creates a type safe tick value instance
func Num[T any](n int) Item[T] {
	return Item[T]{
		N:    n,
		Disc: enums.ItemDiscNumeric,
	}
}

// SendItems is a utility function that sends a list of items and indicates a
// strategy on whether to close the channel once the function completes.
func SendItems[T any](ctx context.Context,
	ch chan<- Item[T], strategy CloseChannelStrategy, items ...any,
) {
	if strategy == CloseChannel {
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

// IsCh checks if an item is an error.
func (i Item[T]) IsCh() bool {
	return (i.Disc & enums.ItemDiscChan) > 0
}

// IsError checks if an item is an error.
func (i Item[T]) IsError() bool {
	return (i.Disc & enums.ItemDiscError) > 0
}

// IsTick checks if an item is a tick instance.
func (i Item[T]) IsTick() bool {
	return (i.Disc & enums.ItemDiscPulse) > 0
}

// IsTickValue checks if an item is a tick instance.
func (i Item[T]) IsTickValue() bool {
	return (i.Disc & enums.ItemDiscTickValue) > 0
}

// IsTickValue checks if an item is a tick instance.
func (i Item[T]) IsNumeric() bool {
	return (i.Disc & enums.ItemDiscNumeric) > 0
}

// SendBlocking sends an item and blocks until it is sent.
func (i Item[T]) SendBlocking(ch chan<- Item[T]) {
	ch <- i
}

// SendContext sends an item and blocks until it is sent or a context canceled.
// It returns a boolean to indicate whether the item was sent.
func (i Item[T]) SendContext(ctx context.Context, ch chan<- Item[T]) bool {
	select {
	case <-ctx.Done(): // Context's done channel has the highest priority
		return false
	default:
		select {
		case <-ctx.Done():
			return false
		case ch <- i:
			return true
		}
	}
}

func (i Item[T]) SendOpContext(ctx context.Context, ch any) bool { // Item[operator[T]]
	_ = ctx
	_ = ch

	panic("SendOpContext: NOT-IMPL")
}

// SendNonBlocking sends an item without blocking.
// It returns a boolean to indicate whether the item was sent.
func (i Item[T]) SendNonBlocking(ch chan<- Item[T]) bool {
	select {
	default:
		return false
	case ch <- i:
		return true
	}
}
