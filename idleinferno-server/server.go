package main

import (
	"log"

	"github.com/kvitebjorn/idleinferno/internal/db"
	"github.com/kvitebjorn/idleinferno/internal/db/sqlite"
	"github.com/kvitebjorn/idleinferno/internal/game"
	"github.com/kvitebjorn/idleinferno/internal/game/model"
)

type Server struct {
	db db.Database
}

func (s *Server) Run() {
	log.Println("Initializing database...")
	s.db = &sqlite.Sqlite{}
	s.db.Init()
	log.Println("Database initialized successfully!")

	log.Println("Starting idleinferno...")
	log.Println("Initializing world...")
	game := game.Game{World: s.initWorld()}
	log.Println("World initialized successfully!")

	log.Println("Running main game loop...")
	game.Run()
	log.Println("Main game loop exited.")

	log.Println("Saving the world...")
	s.saveWorld(game.World)
	log.Println("World saved!")

	err := s.db.Close()
	if err != nil {
		log.Fatalln("Error closing database: ", err.Error())
	}
	log.Println("Exiting idleinferno.")
}

func (s *Server) initWorld() *model.World {
	players := s.db.ReadPlayers()
	items := s.db.ReadItems()

	world := &model.World{}

	for _, p := range players {
		world.Players = append(world.Players, p)
		world.PlayerGrid[p.Location.Y][p.Location.X] = p
	}
	for _, i := range items {
		world.Items = append(world.Items, i)
		world.ItemGrid[i.Location.Y][i.Location.X] = i
	}

	log.Println(world.ToString())
	return world
}

func (s *Server) saveWorld(world *model.World) {
	for _, player := range world.Players {
		_ = s.db.UpdatePlayer(player)
	}

	// TODO: save the rest of the state
	return
}
