package boost

import (
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/samber/lo"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

type WaitGroupName string
type GoRoutineName string
type GoRoutineID string
type namesCollection map[GoRoutineName]string

//
// We need to use variadic parameter list in the methods because of go's lack of
// overloading methods; we need to support calls to wait group methods with or
// without the go routine name behind the same methods name, ie Add needs to be
// able to be invoked either way.

// AnnotatedWgAdder is the interface that is a restricted view of a wait group
// that only allows adding to the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AnnotatedWgAdder interface {
	Add(delta int, name ...GoRoutineName)
}

// AnnotatedWgQuitter is the interface that is a restricted view of a wait group
// that only allows Done signalling on the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AnnotatedWgQuitter interface {
	Done(name ...GoRoutineName)
}

// AnnotatedWgAQ is the interface that is a restricted view of a wait group
// that allows adding to the wait group and Done signalling with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AnnotatedWgAQ interface {
	AnnotatedWgAdder
	AnnotatedWgQuitter
}

// AnnotatedWgWaiter is the interface that is a restricted view of a wait group
// that only allows waiting on the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AnnotatedWgWaiter interface {
	Wait(name ...GoRoutineName)
}

// AnnotatedWgCounter is the interface that is a restricted view of a wait group
// that only allows querying the wait group count. This interface
// can be acquired from the wait group using a standard interface type query.
type AnnotatedWgCounter interface {
	Count() int
}

// WaitGroupAn the extended WaitGroup
type WaitGroupAn interface {
	AnnotatedWgAdder
	AnnotatedWgQuitter
	AnnotatedWgWaiter
	AnnotatedWgCounter
}

type waitGroupAnImpl struct {
	counter       int32
	names         namesCollection
	waitGroupName string
	logger        *slog.Logger
}

func (a *waitGroupAnImpl) Add(delta int, name ...GoRoutineName) {
	atomic.AddInt32(&a.counter, int32(delta))

	if len(name) > 0 {
		a.names[name[0]] = "foo"

		a.indicate(string(name[0]), "Add")
	}
}

func (a *waitGroupAnImpl) Done(name ...GoRoutineName) {
	atomic.AddInt32(&a.counter, int32(-1))

	if len(name) > 0 {
		delete(a.names, name[0])

		a.indicate(string(name[0]), "Done")
	}
}

func (a *waitGroupAnImpl) Wait(name ...GoRoutineName) {
	if len(name) > 0 {
		a.indicate(string(name[0]), "Wait")
	}
}

func (a *waitGroupAnImpl) indicate(name, op string) {
	a.logger.Debug(
		"WaitGroupAssister",
		slog.String("wg-name", a.waitGroupName),
		slog.String("op", op),
		slog.String("name", name),
		slog.Int("counter", int(a.counter)),
		slog.String("running", a.running()),
	)
}

func (a *waitGroupAnImpl) running() string {
	runners := lo.Map(lo.Keys(a.names), func(item GoRoutineName, _ int) string {
		return string(item)
	})

	return strings.Join(runners, "/")
}

// AnnotatedWaitGroup is a wrapper around the standard WaitGroup that
// provides annotations to wait group operations that can assist in
// diagnosing concurrency issues.
type AnnotatedWaitGroup struct {
	wg        sync.WaitGroup
	assistant waitGroupAnImpl
}

// NewAnnotatedWaitGroup creates a new AnnotatedWaitGroup instance containing
// the core WaitGroup instance.
func NewAnnotatedWaitGroup(name string, log ...*slog.Logger) WaitGroupAn {
	logger := lo.TernaryF(len(log) > 0,
		func() *slog.Logger {
			return log[0]
		},
		func() *slog.Logger {
			return slog.New(zapslog.NewHandler(
				zapcore.NewNopCore(), nil),
			)
		},
	)

	return &AnnotatedWaitGroup{
		assistant: waitGroupAnImpl{
			waitGroupName: name,
			names:         make(namesCollection),
			logger:        logger,
		},
	}
}

func (d *AnnotatedWaitGroup) atomic(operation func()) {
	operation()
}

// Add wraps the standard WaitGroup Add operation with the addition of
// being able to associate a go routine (identified by a client provided
// name) with the Add request.
func (d *AnnotatedWaitGroup) Add(delta int, name ...GoRoutineName) {
	d.atomic(func() {
		d.assistant.Add(delta, name...)
		d.wg.Add(delta)
	})
}

// Done wraps the standard WaitGroup Done operation with the addition of
// being able to associate a go routine (identified by a client provided
// name) with the Done request.
func (d *AnnotatedWaitGroup) Done(name ...GoRoutineName) {
	d.atomic(func() {
		d.assistant.Done(name...)
		d.wg.Done()
	})
}

// Wait wraps the standard WaitGroup Wait operation with the addition of
// being able to associate a go routine (identified by a client provided
// name) with the Wait request.
func (d *AnnotatedWaitGroup) Wait(name ...GoRoutineName) {
	d.atomic(func() {
		d.assistant.Wait(name...)
		d.wg.Wait()
	})
}

func (d *AnnotatedWaitGroup) Count() int {
	return int(d.assistant.counter)
}
