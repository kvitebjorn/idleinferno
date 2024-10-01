package model

type ItemClass int

const (
	Head ItemClass = iota + 1
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
	ItemLevel uint
	Class     ItemClass
}
