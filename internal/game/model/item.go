package model

import "fmt"

type ItemClass int

var classToString = []string{
	"helmet",
	"chest plate",
	"grieves",
	"vambraces",
	"gloves",
	"boots",
	"necklace",
	"ring",
	"weapon", // TODO: make this variable: swords, axes, etc.
}

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
	class := classToString[i.Class]
	return fmt.Sprintf("%s, the level %d %s", i.Name, i.ItemLevel, class)
}

func createItem(itemClass ItemClass, itemLevel int) *Item {
	name := "TODO"
	return &Item{
		Name:      name,
		Class:     ItemClass(itemClass),
		ItemLevel: itemLevel}
}
