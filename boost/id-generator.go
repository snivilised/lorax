package boost

import (
	"fmt"
)

type Sequential struct {
	Format string
	id     int
}

func (g *Sequential) Generate() string {
	g.id++

	return fmt.Sprintf(g.Format, g.id)
}
