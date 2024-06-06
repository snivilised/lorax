package boost

import (
	"sync"
	"sync/atomic"
)

type (
	basePool[I, O any] struct {
		wg         *sync.WaitGroup
		sequence   int32
		inputDupCh *Duplex[I]
		oi         *outputInfo[O]
		ending     bool
	}
)

func (p *basePool[I, O]) next() int32 {
	return atomic.AddInt32(&p.sequence, int32(1))
}

// Observe
func (p *basePool[I, O]) Observe() JobOutputStreamR[O] {
	return p.oi.outputDupCh.ReaderCh
}

// CancelCh
func (p *basePool[I, O]) CancelCh() CancelStreamR {
	if p.oi != nil {
		return p.oi.cancelDupCh.ReaderCh
	}

	return nil
}
