package rx

import (
	"context"
	"time"
)

// Infinite represents an infinite wait time
var Infinite int64 = -1

// Duration represents a duration
type Duration interface {
	duration() time.Duration
}

type duration struct {
	d time.Duration
}

func (d *duration) duration() time.Duration {
	return d.d
}

// WithDuration is a duration option
func WithDuration(d time.Duration) Duration {
	return &duration{
		d: d,
	}
}

type causalityDuration struct {
	fs []execution
}

type execution struct {
	f      func()
	isTick bool
}

func timeCausality[T any](elems ...any) (context.Context, Observable[T], Duration) {
	ch := make(chan Item[T], 1)
	fs := make([]execution, len(elems)+1)
	ctx, cancel := context.WithCancel(context.Background())

	for i, elem := range elems {
		i := i
		elem := elem

		if el, ok := elem.(Item[T]); ok && el.IsTick() {
			fs[i] = execution{
				f:      func() {},
				isTick: true,
			}
		} else {
			switch elem := elem.(type) {
			case Item[T]:
				fs[i] = execution{
					f: func() {
						ch <- elem
					},
				}

			case error:
				fs[i] = execution{
					f: func() {
						ch <- Error[T](elem)
					},
				}

			case T:
				fs[i] = execution{
					f: func() {
						ch <- Of(elem)
					},
				}
			}
		}
	}

	fs[len(elems)] = execution{
		f: func() {
			cancel()
		},
		isTick: false,
	}

	return ctx, FromChannel(ch), &causalityDuration{fs: fs}
}

func (d *causalityDuration) duration() time.Duration {
	pop := d.fs[0]
	pop.f()

	d.fs = d.fs[1:]

	if pop.isTick {
		return time.Nanosecond
	}

	return time.Minute
}

// type mockDuration struct {
// 	mock.Mock
// }

// func (m *mockDuration) duration() time.Duration {
// 	args := m.Called()
// 	return args.Get(0).(time.Duration)
// }
