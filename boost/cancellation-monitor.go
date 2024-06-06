package boost

import (
	"context"
	"sync"
)

// StartCancellationMonitor
func StartCancellationMonitor(ctx context.Context,
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	cancelCh CancelStreamR,
	on OnCancel,
) {
	wg.Add(1)
	go func(ctx context.Context,
		cancel context.CancelFunc,
		wg *sync.WaitGroup,
		cancelCh CancelStreamR,
	) {
		defer wg.Done()

		select {
		case <-cancelCh:
			on()
			cancel()
		case <-ctx.Done():
		}
	}(ctx, cancel, wg, cancelCh)
}
