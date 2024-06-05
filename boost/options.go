package boost

import (
	"time"

	"github.com/samber/lo"
)

// withDefaults prepends boost withDefaults to the sequence of options
func withDefaults(options ...Option) []Option {
	const (
		noDefaults = 1
	)
	o := make([]Option, 0, len(options)+noDefaults)
	o = append(o, WithGenerator(&Sequential{
		Format: "ID:%v",
	}))
	o = append(o, options...)

	return o
}

func GetValidatedCheckCloseInterval(o *Options) time.Duration {
	return lo.TernaryF(
		o.Output != nil && o.Output.CheckCloseInterval > minimumCheckCloseInterval,
		func() time.Duration {
			return o.Output.CheckCloseInterval
		},
		func() time.Duration {
			return minimumCheckCloseInterval
		},
	)
}
