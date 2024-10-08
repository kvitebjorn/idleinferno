package game

import (
	"time"

	"github.com/kvitebjorn/idleinferno/internal/game/model"
)

type Game struct {
	World *model.World
}

func (g *Game) Run(saveFn func(world *model.World)) {
	ticker := time.NewTicker(60 * time.Second)
	quit := make(chan struct{})
	for {
		select {

		case <-ticker.C:
			g.tick()
			saveFn(g.World)

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (g *Game) tick() {
	g.World.Wander()
	g.World.Scavenge()
	g.World.Arena()
	g.World.Revelation()
	return
}
