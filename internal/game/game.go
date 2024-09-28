package game

import (
	"log"
	"time"

	"github.com/kvitebjorn/idleinferno/internal/game/model"
)

type Game struct {
	World *model.World
}

func (g *Game) Run() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {

			case <-ticker.C:
				// TODO: for debugging only, remove later
				log.Println(g.World.ToString())

				g.tick()

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// TODO: create a signal for when to close(quit)
	time.Sleep(25 * time.Second)
	close(quit)
}

func (g *Game) tick() {
	g.World.Walk()
	return
}
