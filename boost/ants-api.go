package boost

import "github.com/snivilised/lorax/internal/ants"

type (
	IDGenerator = ants.IDGenerator
	InputParam  = ants.InputParam
	Option      = ants.Option
	Options     = ants.Options
	PoolFunc    = ants.PoolFunc
	TaskFunc    = ants.TaskFunc
)

var (
	WithDisablePurge     = ants.WithDisablePurge
	WithExpiryDuration   = ants.WithExpiryDuration
	WithGenerator        = ants.WithGenerator
	WithInput            = ants.WithInput
	WithMaxBlockingTasks = ants.WithMaxBlockingTasks
	WithNonblocking      = ants.WithNonblocking
	WithOptions          = ants.WithOptions
	WithOutput           = ants.WithOutput
	WithPanicHandler     = ants.WithPanicHandler
	WithPreAlloc         = ants.WithPreAlloc
)
