package ants_test

import (
	"runtime"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ok
	"github.com/snivilised/lorax/boost"
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

func demoPoolFunc(input boost.InputParam) {
	n, _ := input.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

var stopLongRunningFunc int32

func longRunningFunc() {
	for atomic.LoadInt32(&stopLongRunningFunc) == 0 {
		runtime.Gosched()
	}
}

func ShowMemStats() {
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	GinkgoWriter.Printf("memory usage:%d MB", curMem)
}
