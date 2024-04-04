package rx

import (
	"context"
	"reflect"
	"time"
)

type (
	// Item is a wrapper having either a value, error or channel.
	//
	Item[T any] struct {
		V T
		E error
		//
		C       chan<- Item[T]
		tick    bool
		tickV   bool
		numeric bool
		TV      int
		N       int
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
	return Item[T]{V: v}
}

// Ch creates an item from a channel
func Ch[T any](ch any) Item[T] {
	if c, ok := ch.(chan<- Item[T]); ok {
		return Item[T]{C: c}
	}

	panic("temp: invalid ch type")
}

// Tick creates a type safe tick instance
func Tick[T any]() Item[T] {
	return Item[T]{tick: true}
}

// Tv creates a type safe tick value instance
func Tv[T any](tv int) Item[T] {
	return Item[T]{
		TV:    tv,
		tickV: true,
	}
}

// Num creates a type safe tick value instance
func Num[T any](n int) Item[T] {
	return Item[T]{
		N:       n,
		numeric: true,
	}
}

// Error creates an item from an error.
func Error[T any](err error) Item[T] {
	return Item[T]{E: err}
}

// SendItems is an utility function that send a list of items and indicate a
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
	for _, current := range items {
		switch item := current.(type) {
		default:
			rt := reflect.TypeOf(item)

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
				inCh := reflect.ValueOf(current)

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

			case reflect.Slice:
				s := reflect.ValueOf(current)

				for i := 0; i < s.Len(); i++ {
					send(ctx, ch, s.Index(i).Interface())
				}
			}

		case error:
			Error[T](item).SendContext(ctx, ch)
		}
	}
}

// IsCh checks if an item is an error.
func (i Item[T]) IsCh() bool {
	return i.C != nil
}

// IsError checks if an item is an error.
func (i Item[T]) IsError() bool {
	return i.E != nil
}

// IsTick checks if an item is a tick instance.
func (i Item[T]) IsTick() bool {
	return i.tick
}

// IsTickValue checks if an item is a tick instance.
func (i Item[T]) IsTickValue() bool {
	return i.tickV
}

// IsTickValue checks if an item is a tick instance.
func (i Item[T]) IsNumeric() bool {
	return i.numeric
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
