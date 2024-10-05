package model

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"sync"
)

// Probably only the 3 of us playing, so...
// I guess each circle is only 1 array + purgatory = 10
const WorldSize int = 10

type Coordinates struct {
	X int
	Y int
}

var Up Coordinates = Coordinates{X: 0, Y: -1}
var Down Coordinates = Coordinates{X: 0, Y: 1}
var Left Coordinates = Coordinates{X: -1, Y: 0}
var Right Coordinates = Coordinates{X: 1, Y: 0}

type World struct {
	Players []*Player
	Grid    [WorldSize][WorldSize]*Player

	mut sync.Mutex
}

func (w *World) Login(player *Player) (*Player, error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	if w.Grid[player.Location.Y][player.Location.X] == nil {
		w.Grid[player.Location.Y][player.Location.X] = player
		w.Players = append(w.Players, player)
		return player, nil
	}

	for i := 0; i < WorldSize; i++ {
		for j := 0; j < WorldSize; j++ {
			if w.Grid[i][j] == nil {
				w.Grid[i][j] = player
				player.Location.X = j
				player.Location.Y = i
				w.Players = append(w.Players, player)
				return player, nil
			}
		}
	}

	return nil, errors.New("Unable to place player in world.")
}

func (w *World) Logout(player *Player) {
	w.mut.Lock()
	defer w.mut.Unlock()

	newPlayers := make([]*Player, 0, len(w.Players)-1)
	for _, p := range w.Players {
		if p.Name != player.Name {
			newPlayers = append(newPlayers, p)
		}
	}
	w.Players = newPlayers
	w.Grid[player.Location.Y][player.Location.X] = nil
}

func (w *World) Wander() {
	w.mut.Lock()
	defer w.mut.Unlock()

	for _, player := range w.Players {
		player.Stats.IncrementXp()
		emptyNeighborCoords := w.getEmptyNeighborCoords(player.Location)
		emptyNeighborCoordsLen := len(emptyNeighborCoords)
		if emptyNeighborCoordsLen == 0 {
			continue
		}
		destCoords := emptyNeighborCoords[rand.IntN(emptyNeighborCoordsLen)]
		w.Grid[player.Location.Y][player.Location.X] = nil
		w.Grid[destCoords.Y][destCoords.X] = player
		player.Location.X = destCoords.X
		player.Location.Y = destCoords.Y
	}
}

func (w *World) Scavenge() {
	w.mut.Lock()
	defer w.mut.Unlock()

	for _, player := range w.Players {
		player.FindItem()
	}
}

func (w *World) Arena() {
	w.mut.Lock()
	defer w.mut.Unlock()

	// TODO: can i cache this somehow? kind of want the anti-walk data
	//       but we'd have to fight before we walk...
	//       meh
	alreadyFought := make(map[string]bool)
	for _, player := range w.Players {
		if ok := alreadyFought[player.Name]; ok {
			continue
		}
		neighborCoords := w.getOccupiedNeighborCoords(player.Location)
		neighborCoordsLen := len(neighborCoords)
		if neighborCoordsLen == 0 {
			continue
		}
		opponentCoords := neighborCoords[rand.IntN(neighborCoordsLen)]
		opponent := w.Grid[opponentCoords.Y][opponentCoords.X]

		if player.ItemLevel() == 0 || opponent.ItemLevel() == 0 {
			continue
		}

		playerRoll := rand.IntN(player.ItemLevel())
		opponentRoll := rand.IntN(opponent.ItemLevel())
		result := "won"
		if opponentRoll >= playerRoll {
			result = "lost"
			opponent.Stats.IncrementXp()
			player.Stats.DecrementXp()
		} else {
			opponent.Stats.DecrementXp()
			player.Stats.IncrementXp()
		}
		combatMsg := fmt.Sprintf("%s (%d) challenged %s (%d) to a fight and %s!",
			player.Name,
			playerRoll,
			opponent.Name,
			opponentRoll,
			result)
		log.Println(combatMsg)

		alreadyFought[player.Name] = true
		alreadyFought[opponent.Name] = true
	}
}

func (w *World) ToString() string {
	w.mut.Lock()
	defer w.mut.Unlock()

	infernoArt := []string{
		"       __________________________________  ",
		"      |       Circle 1: Limbo          |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 2: Lust          |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 3: Gluttony       |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 4: Greed          |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 5: Wrath          |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 6: Heresy         |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 7: Violence        |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 8: Fraud          |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |______________________________    |  ",
		"      |       Circle 9: Treachery      |  ",
		"      |            __________           |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |          |          |  ",
		"      |           |__________|          |  ",
		"      |__________________________________|  ",
		"                                            ",
		"____________________________________________",
	}

	// Circle bounds for random placement
	xMin, xMax := 8, 36
	yMin, yMax := 4, 7

	// Player coordinates mapping
	playerCoords := map[string]Coordinates{}

	// Randomly assign coordinates within the circle bounds
	// This is because we technically only are one array,
	// but we have to visually occupy a whole circle/mini-grid
	for _, player := range w.Players {
		layer := player.Location.Y
		x := rand.IntN(xMax-xMin+1) + xMin
		y := rand.IntN(yMax-yMin+1) + (layer * 8) + yMin
		playerCoords[player.Name] = Coordinates{X: x, Y: y}
	}

	// Place players on the ASCII art
	for playerName, coord := range playerCoords {
		x := coord.X
		y := coord.Y
		infernoArt[y] = infernoArt[y][:x] + playerName[0:1] + infernoArt[y][x+1:]
	}

	var playerList []string
	for _, player := range w.Players {
		playerList = append(playerList,
			fmt.Sprintf("%s (%d)", player.Name, player.ItemLevel()))
	}

	return strings.Join(infernoArt, "\n") +
		"\n\nPlayers:\n" +
		strings.Join(playerList, "\n")
}

func (w *World) getEmptyNeighborCoords(c *Coordinates) []Coordinates {
	return w.getNeighborCoords(c, false)
}

func (w *World) getOccupiedNeighborCoords(c *Coordinates) []Coordinates {
	return w.getNeighborCoords(c, true)
}

func (w *World) getNeighborCoords(c *Coordinates, lookingForPlayers bool) []Coordinates {
	coords := make([]Coordinates, 0)
	dirs := [...]Coordinates{Up, Down, Left, Right}
	for _, dir := range dirs {
		thisY := c.Y + dir.Y
		thisX := c.X + dir.X
		if thisY < WorldSize &&
			thisY >= 0 &&
			thisX < WorldSize &&
			thisX >= 0 &&
			((lookingForPlayers && w.Grid[thisY][thisX] != nil) ||
				(!lookingForPlayers && w.Grid[thisY][thisX] == nil)) {
			coords = append(coords, Coordinates{X: thisX, Y: thisY})
		}
	}
	return coords
}
