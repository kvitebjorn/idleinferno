package db

import "github.com/kvitebjorn/idleinferno/internal/game/model"

type Database interface {
	Init()
	Close() error

	CreatePlayer(*model.User) *model.Player
	ReadPlayer(name string) *model.Player
	ReadPlayers() []*model.Player
	UpdatePlayer(*model.Player) int64
	DeletePlayer(guid string)

	CreateItem(*model.Item) *model.Item
	ReadItem(guid string) *model.Item
	ReadItems() []*model.Item
	UpdateItem(*model.Item) int64
	DeleteItem(guid string)

	ReadUser(name string) *model.User
	ReadUserByEmail(email string) *model.User
	UpdateUserOnline(name string) error
	UpdateUserOffline(name string) error
}
