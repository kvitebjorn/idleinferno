package model

import (
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"text/tabwriter"
)

type Player struct {
	Id        string
	Name      string
	Class     string
	Stats     *Stats
	Inventory [9]*Item
	Location  *Coordinates
}

type User struct {
	Name     string
	Password string
	Email    string
	Class    string
	Online   bool
}

func (p Player) ItemLevel() int {
	sum := 0
	for _, i := range p.Inventory {
		if i == nil {
			continue
		}
		sum += i.ItemLevel
	}
	return sum
}

/*
There is about a 10% chance to find an item per turn.

If an item is found with the 10% base chance, for a level 20 player,

	  the percentage chance to then find items of the following levels is:

		Level 1: 7.97%
		Level 5: 6.52%
		Level 10: 4.71%
		Level 15: 2.90%
		Level 20: 1.09%
*/
func (p *Player) FindItem() {
	// Base chance of finding an item
	playerRollToFindTheItem := float64(p.Stats.Level()+1) / 200.0

	// Random chance to find an item
	chanceToFindTheItem := rand.Float64()

	rollMsg := fmt.Sprintf("%s rolled a %d to find an item (%d)",
		p.Name, int(playerRollToFindTheItem*100), int(chanceToFindTheItem*100))
	fmt.Println(rollMsg)
	if chanceToFindTheItem > playerRollToFindTheItem {
		return
	}

	// Item class
	itemClass := rand.IntN(9)

	// Randomly determine item level with bias towards lower levels
	maxLevel := p.Stats.Level() + 2
	itemLevel := weightedRandomItemLevel(maxLevel)

	// Item level adjustment
	k := 2.0 // the scale factor for difficulty
	itemLevelChance := 1.0 / float64((itemLevel+1)^int(k))

	// Calculate final chance of getting this item level
	finalChance := playerRollToFindTheItem * itemLevelChance

	if rand.Float64() > finalChance {
		fmt.Println(p.Name, "found no items.")
	}

	// Check if the found item is worse than existing one
	if p.Inventory[itemClass] != nil && p.Inventory[itemClass].ItemLevel > itemLevel {
		fmt.Println(p.Name,
			"found a new item, but it's worse than their",
			p.Inventory[itemClass].ToString())
		return
	}

	// Create and add the new item
	newItem := createItem(ItemClass(itemClass), itemLevel)
	newItem.Player = p.Name
	p.Inventory[itemClass] = newItem
	log.Println(p.Name, "equipped a", newItem.ToString())
}

// weightedRandomItemLevel generates a random item level with bias towards lower levels
func weightedRandomItemLevel(maxLevel int) int {
	// Create a weighted distribution
	weights := make([]float64, maxLevel+1)
	totalWeight := 0.0

	for i := 0; i <= maxLevel; i++ {
		// Higher weight for lower levels
		weights[i] = float64(maxLevel - i + 1) // e.g., maxLevel=3 -> weights = [4, 3, 2, 1]
		totalWeight += weights[i]
	}

	// Randomly select an item level based on weights
	roll := rand.Float64() * totalWeight
	for level, weight := range weights {
		if roll < weight {
			return level
		}
		roll -= weight
	}

	// Fallback (should not be reached)
	return maxLevel
}

func (p *Player) ToString() string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 4, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "Name: %s\n", p.Name)
	fmt.Fprintf(tw, "Class: %s\n", p.Class)
	fmt.Fprintf(tw, "Item level: %d\n", p.ItemLevel())
	fmt.Fprintf(tw, "Experience: %d\n", p.Stats.Xp)
	fmt.Fprintf(tw, "Level: %d\n", p.Stats.Level())
	fmt.Fprintf(tw, "Next level: %d\n", p.Stats.UntilNextLevel())
	fmt.Fprintf(tw, "Location: (%d,%d)\n", p.Location.X, p.Location.Y)
	fmt.Fprintf(tw, "Id: %s\n", p.Id)
	fmt.Fprintf(tw, "Created: %s\n", p.Stats.Created)
	fmt.Fprintf(tw, "Inventory:\n")
	for _, i := range p.Inventory {
		if i == nil {
			continue
		}
		fmt.Fprintf(tw, "%s (%d)\n", i.Name, i.ItemLevel)
	}
	tw.Flush()
	return sb.String()
}
