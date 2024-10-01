package model

type Player struct {
	Id        string
	Name      string
	Class     string
	Stats     *Stats
	Inventory *Inventory
	Location  *Coordinates
}

type User struct {
	Name     string
	Password string
	Email    string
	Class    string
	Online   bool
}
