package boost

import "github.com/snivilised/lorax/internal/ants"

type (
	Option = ants.Option
)

var (
	WithDisablePurge     = ants.WithDisablePurge
	WithExpiryDuration   = ants.WithExpiryDuration
	WithMaxBlockingTasks = ants.WithMaxBlockingTasks
	WithNonblocking      = ants.WithNonblocking
	WithOptions          = ants.WithOptions
	WithPanicHandler     = ants.WithPanicHandler
	WithPreAlloc         = ants.WithPreAlloc
)
