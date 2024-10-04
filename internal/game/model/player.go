package model

import (
	"log"
	"math/rand/v2"
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

func (p *Player) AcquireItem() {
	itemClass := rand.IntN(9)
	itemLevel := rand.IntN(p.Stats.Level() + 2)
	if p.Inventory[itemClass] != nil &&
		p.Inventory[itemClass].ItemLevel > itemLevel {
		log.Println(p.Name,
			"found a new item, but it's worse than their",
			p.Inventory[itemClass].ToString())
		return
	}
	// TODO: enable this when we have Stats:
	//  p.Stats.ItemLevel += itemLevel - p.Inventory[itemClass].ItemLevel
	newItem := createItem(ItemClass(itemClass), itemLevel)
	newItem.Player = p.Name
	p.Inventory[itemClass] = newItem
	log.Println(p.Name, "acquired", newItem.ToString())
}
