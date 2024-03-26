package rx

import (
	"context"
	"time"
)

type (
	// Item is a wrapper having either a value or an error.
	//
	Item[I any] struct {
		V I
		E error
	}

	// TimestampItem attach a timestamp to an item.
	//
	TimestampItem[I any] struct {
		Timestamp time.Time
		V         I
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
func Of[I any](v I) Item[I] {
	return Item[I]{V: v}
}

// Error creates an item from an error.
func Error[I any](err error) Item[I] {
	return Item[I]{E: err}
}

// SendItems is an utility function that send a list of items and indicate a
// strategy on whether to close the channel once the function completes.
// This method has been derived from the original SendItems.
// (does not support channels or slice)
func SendItems[I any](ctx context.Context,
	ch chan<- Item[I], strategy CloseChannelStrategy, items ...Item[I],
) {
	if strategy == CloseChannel {
		defer close(ch)
	}

	sendItems(ctx, ch, items...)
}

func sendItems[I any](ctx context.Context, ch chan<- Item[I], items ...Item[I]) {
	for _, item := range items {
		item.SendContext(ctx, ch)
	}
}

// IsError checks if an item is an error.
func (i Item[I]) IsError() bool {
	return i.E != nil
}

// SendBlocking sends an item and blocks until it is sent.
func (i Item[I]) SendBlocking(ch chan<- Item[I]) {
	ch <- i
}

// SendContext sends an item and blocks until it is sent or a context canceled.
// It returns a boolean to indicate whether the item was sent.
func (i Item[I]) SendContext(ctx context.Context, ch chan<- Item[I]) bool {
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

func (i Item[I]) SendOpContext(ctx context.Context, ch any) bool { // Item[operator[I]]
	_ = ctx
	_ = ch

	panic("SendOpContext: NOT-IMPL")
}

// SendNonBlocking sends an item without blocking.
// It returns a boolean to indicate whether the item was sent.
func (i Item[I]) SendNonBlocking(ch chan<- Item[I]) bool {
	select {
	default:
		return false
	case ch <- i:
		return true
	}
}
