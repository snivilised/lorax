package boost_test

import (
	"runtime"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ok
	"github.com/snivilised/lorax/internal/ants"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)

const (
	Param    = 100
	AntsSize = 1000
	TestSize = 10000
	n        = 100000
	PoolSize = 10
)

const (
	RunTimes           = 1e6
	PoolCap            = 5e4
	BenchParam         = 10
	DefaultExpiredTime = 10 * time.Second
)

var curMem uint64

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

func demoPoolFunc(inputCh ants.InputParam) {
	n, _ := inputCh.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

var stopLongRunningFunc int32

func longRunningFunc() {
	for atomic.LoadInt32(&stopLongRunningFunc) == 0 {
		runtime.Gosched()
	}
}

var stopLongRunningPoolFunc int32

func longRunningPoolFunc(arg ants.InputParam) {
	if ch, ok := arg.(chan struct{}); ok {
		<-ch
		return
	}
	for atomic.LoadInt32(&stopLongRunningPoolFunc) == 0 {
		runtime.Gosched()
	}
}

func ShowMemStats() {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	GinkgoWriter.Printf("memory usage:%d MB", curMem)
}
