package boost

import (
	"fmt"
	"sync/atomic"
)

type Sequential struct {
	Format string
	id     int32
}

func (g *Sequential) Generate() string {
	n := atomic.AddInt32(&g.id, int32(1))

	return fmt.Sprintf(g.Format, n)
}
