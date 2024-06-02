package boost

import (
	"sync"
	"sync/atomic"
)

type (
	basePool[O any] struct {
		wg          *sync.WaitGroup
		sequence    int32
		outputDupCh *Duplex[JobOutput[O]]
		ending      bool
	}
)

func (p *basePool[O]) next() int32 {
	return atomic.AddInt32(&p.sequence, int32(1))
}

func (p *basePool[O]) Observe() JobOutputStreamR[O] {
	return p.outputDupCh.ReaderCh
}
