package boost

import (
	"sync"
	"sync/atomic"
)

type (
	basePool[I, O any] struct {
		wg          *sync.WaitGroup
		sequence    int32
		inputDupCh  *Duplex[I]
		outputDupCh *Duplex[JobOutput[O]]
		ending      bool
	}
)

func (p *basePool[I, O]) next() int32 {
	return atomic.AddInt32(&p.sequence, int32(1))
}

func (p *basePool[I, O]) Observe() JobOutputStreamR[O] {
	return p.outputDupCh.ReaderCh
}
