package model

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"sync"
	"text/tabwriter"
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

	newPlayers := make([]*Player, 0)
	for _, p := range w.Players {
		if p.Name == player.Name {
			continue
		}
		newPlayers = append(newPlayers, p)
	}
	w.Players = newPlayers
	w.Grid[player.Location.Y][player.Location.X] = nil
}

func (w *World) Wander() {
	w.mut.Lock()
	defer w.mut.Unlock()

	for _, player := range w.Players {
		player.Stats.Xp += 1
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
		chance := rand.IntN(int(player.Stats.Level() + 2))
		log.Println(player.Name, "rolled a", chance)
		if chance > int(player.Stats.Level()/2) {
			player.AcquireItem()
		}
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

		playerRoll := rand.IntN(player.ItemLevel())
		opponentRoll := rand.IntN(opponent.ItemLevel())
		result := "won"
		if opponentRoll >= playerRoll {
			result = "lost"
		}
		combatMsg := fmt.Sprintf("%s (%d) challenged %s (%d) to a fight and %s!",
			player.Name,
			playerRoll,
			opponent.Name,
			opponentRoll,
			result)
		log.Println(combatMsg)

		// TODO: add/subtract xp from each player
		// TODO: save a record of this fight for each player, and its result

		alreadyFought[player.Name] = true
		alreadyFought[opponent.Name] = true
	}
}

func (w *World) ToString() string {
	w.mut.Lock()
	defer w.mut.Unlock()

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 4, 1, 1, ' ', 0)
	fmt.Fprintln(tw, "Current state of the inferno:")
	fmt.Fprintln(tw, "")
	for y := 0; y < WorldSize; y++ {
		fmt.Fprint(tw, "|\t")
		for x := 0; x < WorldSize; x++ {
			if w.Grid[y][x] != nil {
				fmt.Fprint(tw, w.Grid[y][x].Name+"\t")
			} else {
				fmt.Fprint(tw, "_\t")
			}
		}
		fmt.Fprintln(tw, "|")
	}
	tw.Flush()

	sb.WriteString("\n")
	sb.WriteString("Players:\n")
	for _, p := range w.Players {
		playerDescription := fmt.Sprintf("%s (%d)\n", p.Name, p.ItemLevel())
		sb.WriteString(playerDescription)
	}

	sb.WriteString("\n")

	return sb.String()
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
