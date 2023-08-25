package async

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type WaitGroupName string
type GoRoutineName = string
type GoRoutineID string
type namesCollection map[GoRoutineName]string

func ComposeGoRoutineID(prefix ...string) GoRoutineID {
	id := uuid.NewString()

	return lo.TernaryF(len(prefix) > 0,
		func() GoRoutineID {
			return GoRoutineID(fmt.Sprintf("GR-ID(%v):%v", prefix[0], id))
		},
		func() GoRoutineID {
			return GoRoutineID(fmt.Sprintf("GR-ID:%v", id))
		},
	)
}

// WaitGroupEx the extended WaitGroup
type WaitGroupEx interface {
	Add(delta int, name ...GoRoutineName)
	Done(name ...GoRoutineName)
	Wait(name ...GoRoutineName)
}

// ===> Adder

type AssistedAdder interface {
	Add(delta int, name ...GoRoutineName)
}

// ===> Quitter

type AssistedQuitter interface {
	Done(name ...GoRoutineName)
}

// ===> Waiter

type AssistedWaiter interface {
	Done(name ...GoRoutineName)
}

// ===> Counter
type AssistedCounter interface {
	Count() int
}

type WaitGroupAssister struct {
	counter int32
	names   namesCollection
}

func (a *WaitGroupAssister) Add(delta int, name ...GoRoutineName) {
	a.counter += int32(delta)

	if len(name) > 0 {
		a.names[name[0]] = "foo"

		fmt.Printf("		🟪🟪🟪 [[ WaitGroupAssister.Add ]] - name: '%v' (delta: '%v', count: '%v') (running: '%v')\n",
			name[0], delta, a.counter, a.running(),
		)
	}
}

func (a *WaitGroupAssister) Done(name ...GoRoutineName) {
	a.counter--

	if len(name) > 0 {
		delete(a.names, name[0])

		fmt.Printf("		🔷🔷🔷 [[ WaitGroupAssister.Done ]] - name: '%v' (count: '%v') (running: '%v')\n",
			name[0], a.counter, a.running(),
		)
	}
}

func (a *WaitGroupAssister) Wait(name ...GoRoutineName) {
	if len(name) > 0 {
		fmt.Printf("		🟤🟤🟤 [[ WaitGroupAssister.Wait ]] - name: '%v' (count: '%v') (running: '%v')\n",
			name[0], a.counter, a.running(),
		)
	}
}

func (a *WaitGroupAssister) running() string {
	return strings.Join(lo.Keys(a.names), "/")
}

// You start off with a core instance and from here you can query it to get the
// assistant interfaces ... The WaitGroupCore and the WaitGroupAssister both
// need to know about each other, but making them hold a reference to each other
// will cause a memory leak. To resolve this, there is a release mechanism
// which is automatically enacted after the Wait has become unblocked. The
// consequence of this characteristic is that once a Wait has completed, it
// should not be re-used.
//
// We need to use variadic parameter list in the methods because of go's lack of
// overloading methods; we need to support calls to wait group methods with or
// without the go routine name behind the same methods name, ie Add needs to be
// able to be invoked either way.
// type WaitGroupCore struct {
// 	wg sync.WaitGroup
// }

type AnnotatedWaitGroup struct {
	wg        sync.WaitGroup
	assistant WaitGroupAssister
	mux       sync.Mutex
}

func NewAnnotatedWaitGroup(_ string) *AnnotatedWaitGroup {
	return &AnnotatedWaitGroup{
		assistant: WaitGroupAssister{
			names: make(namesCollection),
		},
	}
}

func (d *AnnotatedWaitGroup) atomic(operation func()) {
	operation()
}

func (d *AnnotatedWaitGroup) Add(delta int, name ...GoRoutineName) {
	d.atomic(func() {
		d.assistant.Add(delta, name...)
		d.wg.Add(delta)
	})
}

func (d *AnnotatedWaitGroup) Done(name ...GoRoutineName) {
	d.atomic(func() {
		d.assistant.Done(name...)
		d.wg.Done()
	})
}

func (d *AnnotatedWaitGroup) Wait(name ...GoRoutineName) {
	// We could make the wait an active operation, that includes
	// creating a new go routine and a channel. The go routine
	// will monitor the state of the wait group, every time
	// either an Add or Done occurs, those send a message to the go
	// routine via this channel. The go routing will simply keep
	// reading the channel and displaying the current count. The
	// active-ness is just a debugging tool, not intended to be
	// used in production as it will be noisy by design either writing
	// to the console or preferably writing to a log.
	//
	d.atomic(func() {
		d.assistant.Wait(name...)
		d.wg.Wait()
	})
}

func (d *AnnotatedWaitGroup) Count() int {
	return int(d.assistant.counter)
}
