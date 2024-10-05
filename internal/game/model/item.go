package model

import (
	"fmt"
)

type ItemClass int

const (
	Head ItemClass = iota
	Torso
	Legs
	Arms
	Gloves
	Boots
	Necklace
	Ring
	Weapon
)

type Item struct {
	Id        string
	Name      string
	ItemLevel int
	Class     ItemClass
	Player    string
}

func (i Item) ToString() string {
	return fmt.Sprintf("level %d %s", i.ItemLevel, i.Name)
}

func createItem(itemClass ItemClass, itemLevel int) *Item {
	name := GetItemName(itemClass)
	return &Item{
		Name:      name,
		Class:     ItemClass(itemClass),
		ItemLevel: itemLevel}
}
