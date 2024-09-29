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
	Email    string
	Password string
}
