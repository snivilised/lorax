package async

import (
	"fmt"
	"strings"
	"sync"

	"github.com/samber/lo"
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

// AssistedAdder is the interface that is a restricted view of a wait group
// that only allows adding to the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AssistedAdder interface {
	Add(delta int, name ...GoRoutineName)
}

// AssistedQuitter is the interface that is a restricted view of a wait group
// that only allows Done signalling on the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AssistedQuitter interface {
	Done(name ...GoRoutineName)
}

// AssistedWaiter is the interface that is a restricted view of a wait group
// that only allows waiting on the wait group with the addition of being
// able to specify the name representing the calling go routine. This interface
// can be acquired from the wait group using a standard interface type query.
type AssistedWaiter interface {
	Wait(name ...GoRoutineName)
}

// WaitGroupEx the extended WaitGroup
type WaitGroupEx interface {
	AssistedAdder
	AssistedQuitter
	AssistedWaiter
}

// AssistedCounter is the interface that is a restricted view of a wait group
// that only allows querying the wait group count. This interface
// can be acquired from the wait group using a standard interface type query.
type AssistedCounter interface {
	Count() int
}

type waitGroupAssister struct {
	counter       int32
	names         namesCollection
	waitGroupName string
}

func (a *waitGroupAssister) Add(delta int, name ...GoRoutineName) {
	a.counter += int32(delta)

	if len(name) > 0 {
		a.names[name[0]] = "foo"

		a.indicate("➕➕➕", string(name[0]), "Add")
	}
}

func (a *waitGroupAssister) Done(name ...GoRoutineName) {
	a.counter--

	if len(name) > 0 {
		delete(a.names, name[0])

		a.indicate("🚩🚩🚩", string(name[0]), "Done")
	}
}

func (a *waitGroupAssister) Wait(name ...GoRoutineName) {
	if len(name) > 0 {
		a.indicate("🧭🧭🧭", string(name[0]), "Wait")
	}
}

func (a *waitGroupAssister) indicate(highlight, name, op string) {
	fmt.Printf(
		"		%v [[ WaitGroupAssister(%v).%v ]] - gr-name: '%v' (count: '%v') (running: '%v')\n",
		highlight, a.waitGroupName, op, name, a.counter, a.running(),
	)
}

func (a *waitGroupAssister) running() string {
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
	assistant waitGroupAssister
	mux       sync.Mutex
}

// NewAnnotatedWaitGroup creates a new AnnotatedWaitGroup instance containing
// the core WaitGroup instance.
func NewAnnotatedWaitGroup(name string) *AnnotatedWaitGroup {
	return &AnnotatedWaitGroup{
		assistant: waitGroupAssister{
			waitGroupName: name,
			names:         make(namesCollection),
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
