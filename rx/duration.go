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

func timeCausality[T any](values ...any) (context.Context, Observable[T], Duration) {
	ch := make(chan Item[T], 1)
	fs := make([]execution, len(values)+1)
	ctx, cancel := context.WithCancel(context.Background())

	for i, value := range values {
		i := i
		value := value

		if el, ok := value.(Item[T]); ok && el.IsTick() {
			fs[i] = execution{
				f:      func() {},
				isTick: true,
			}
		} else {
			switch value := value.(type) {
			case Item[T]:
				fs[i] = execution{
					f: func() {
						ch <- value
					},
				}

			case error:
				fs[i] = execution{
					f: func() {
						ch <- Error[T](value)
					},
				}

			case T:
				fs[i] = execution{
					f: func() {
						ch <- Of(value)
					},
				}
			}
		}
	}

	fs[len(values)] = execution{
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
