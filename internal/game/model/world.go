package model

import (
	"errors"
	"fmt"
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
	Players    []*Player
	PlayerGrid [WorldSize][WorldSize]*Player

	Items    []*Item
	ItemGrid [WorldSize][WorldSize]*Item

	mut sync.Mutex
}

func (w *World) Login(player *Player) (*Player, error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	if w.PlayerGrid[player.Location.Y][player.Location.X] == nil {
		w.PlayerGrid[player.Location.Y][player.Location.X] = player
		w.Players = append(w.Players, player)
		return player, nil
	}

	for i := 0; i < WorldSize; i++ {
		for j := 0; j < WorldSize; j++ {
			if w.PlayerGrid[i][j] == nil {
				w.PlayerGrid[i][j] = player
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
	w.PlayerGrid[player.Location.Y][player.Location.X] = nil
}

func (w *World) Walk() {
	w.mut.Lock()
	defer w.mut.Unlock()
	// TODO: we don't account for items in any way yet
	//       in the future, if landing on an item, potentially equip it
	// Make all players walk 1 in a random direction; only straight lines for now
	for y := 0; y < WorldSize; y++ {
		for x := 0; x < WorldSize; x++ {
			player := w.PlayerGrid[y][x]
			if player == nil {
				continue
			}
			emptyNeighborCoords := w.getEmptyNeighborCoords(player.Location)
			emptyNeighborCoordsLen := len(emptyNeighborCoords)
			destCoords := emptyNeighborCoords[rand.IntN(emptyNeighborCoordsLen)]
			w.PlayerGrid[y][x] = nil
			w.PlayerGrid[destCoords.Y][destCoords.X] = player
			player.Location.X = destCoords.X
			player.Location.Y = destCoords.Y
		}
	}
}

func (w *World) ToString() string {
	w.mut.Lock()
	defer w.mut.Unlock()

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 4, 1, 1, ' ', 0)
	fmt.Fprintln(tw, "Current state of the world:")
	fmt.Fprintln(tw, "")
	for y := 0; y < WorldSize; y++ {
		fmt.Fprint(tw, "|\t")
		for x := 0; x < WorldSize; x++ {
			if w.PlayerGrid[y][x] != nil {
				fmt.Fprint(tw, w.PlayerGrid[y][x].Name+"\t")
			} else if w.ItemGrid[y][x] != nil {
				fmt.Fprint(tw, w.ItemGrid[y][x].Name+"\t")
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
		sb.WriteString(p.Name + "\n")
	}

	sb.WriteString("\n")

	sb.WriteString("Items:\n")
	for _, i := range w.Items {
		sb.WriteString(i.Name + "\n")
	}

	sb.WriteString("\n")

	return sb.String()
}

func (w *World) getEmptyNeighborCoords(c *Coordinates) []Coordinates {
	empty := make([]Coordinates, 0)
	dirs := [...]Coordinates{Up, Down, Left, Right}
	for _, dir := range dirs {
		thisY := c.Y + dir.Y
		thisX := c.X + dir.X
		if thisY < WorldSize &&
			thisY >= 0 &&
			thisX < WorldSize &&
			thisX >= 0 &&
			w.PlayerGrid[thisY][thisX] == nil {
			empty = append(empty, Coordinates{X: thisX, Y: thisY})
		}
	}
	return empty
}
