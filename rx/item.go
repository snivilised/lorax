package rx

import (
	"context"
	"time"
)

type (
	// Item is a wrapper having either a value or an error.
	//
	Item[T any] struct {
		V T
		E error
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

// Error creates an item from an error.
func Error[T any](err error) Item[T] {
	return Item[T]{E: err}
}

// SendItems is an utility function that send a list of items and indicate a
// strategy on whether to close the channel once the function completes.
// This method has been derived from the original SendItems.
// (does not support channels or slice)
func SendItems[T any](ctx context.Context,
	ch chan<- Item[T], strategy CloseChannelStrategy, items ...Item[T],
) {
	if strategy == CloseChannel {
		defer close(ch)
	}

	sendItems(ctx, ch, items...)
}

func sendItems[T any](ctx context.Context, ch chan<- Item[T], items ...Item[T]) {
	for _, item := range items {
		item.SendContext(ctx, ch)
	}
}

// IsError checks if an item is an error.
func (i Item[T]) IsError() bool {
	return i.E != nil
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
