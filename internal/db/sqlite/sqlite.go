package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/kvitebjorn/idleinferno/internal/db/sqlite/queries"
	"github.com/kvitebjorn/idleinferno/internal/game/model"
)

const DatabaseName = "./idleinferno.db"

type Sqlite struct {
	db *sql.DB
}

func (s *Sqlite) Init() {
	createTables := false
	_, err := os.Stat(DatabaseName)
	if err != nil {
		fmt.Println("Creating database...")
		createTables = true
		_, err = os.Create(DatabaseName)
		if err != nil {
			fmt.Println("Failed to create database:", err.Error())
		}
		fmt.Println("Database created successfully!")
	}

	fmt.Println("Connecting to database...")
	db, err := sql.Open("sqlite3", DatabaseName)
	if err != nil {
		log.Fatalln("Error connecting to db:", err)
	}
	fmt.Println("Database connection successful!")

	s.db = db

	if createTables {
		fmt.Println("Creating database tables...")
		_, err = db.Exec(queries.CreatePlayersTableSql)
		_, err = db.Exec(queries.CreateItemsTableSql)
		if err != nil {
			log.Fatalln("Failed to initialize database tables:", err.Error())
		}
		fmt.Println("Database tables created successfully!")
	}

	return
}

func (s *Sqlite) Close() error {
	return s.db.Close()
}

func (s *Sqlite) CreatePlayer(user *model.User) *model.Player {
	stmt, err := s.db.Prepare(queries.CreatePlayerSql)
	checkErr(err)
	defer stmt.Close()

	player := &model.Player{}
	player.Id = uuid.New().String()
	player.Name = user.Name
	player.Class = user.Class

	res, err := stmt.Exec(
		player.Id,
		user.Name,
		user.Email,
		user.Password,
		user.Class)
	checkErr(err)

	_, err = res.RowsAffected()
	checkErr(err)

	return player
}

func (s *Sqlite) ReadPlayer(name string) *model.Player {
	row := s.db.QueryRow(queries.ReadPlayerSql, name)

	player := &model.Player{
		Location: &model.Coordinates{X: 0, Y: 0},
		Stats:    &model.Stats{},
	}

	err := row.Scan(
		&player.Id,
		&player.Name,
		&player.Class,
		&player.Location.X,
		&player.Location.Y,
		&player.Stats.Xp,
		&player.Stats.Created,
		&player.Stats.Online,
	)

	if err != nil {
		return nil
	}

	items := s.ReadItems(player.Name)
	for _, i := range items {
		player.Inventory[i.Class] = i
	}

	return player
}

func (s *Sqlite) ReadPlayers() []*model.Player {
	rows, err := s.db.Query(queries.ReadPlayersSql)
	defer rows.Close()

	checkErr(err)

	players := make([]*model.Player, 0)
	for rows.Next() {
		player := &model.Player{
			Location: &model.Coordinates{X: 0, Y: 0},
			Stats:    &model.Stats{},
		}

		err = rows.Scan(
			&player.Id,
			&player.Name,
			&player.Class,
			&player.Location.X,
			&player.Location.Y,
			&player.Stats.Xp,
			&player.Stats.Created,
			&player.Stats.Online,
		)
		checkErr(err)

		items := s.ReadItems(player.Name)
		for _, i := range items {
			player.Inventory[i.Class] = i
		}

		players = append(players, player)
	}

	err = rows.Err()
	checkErr(err)

	return players
}

func (s *Sqlite) ReadUser(name string) *model.User {
	row := s.db.QueryRow(queries.ReadUserSql, name)

	user := &model.User{}
	err := row.Scan(&user.Name, &user.Password, &user.Online)
	if err != nil {
		return nil
	}

	return user
}

func (s *Sqlite) ReadUserByEmail(email string) *model.User {
	row := s.db.QueryRow(queries.ReadUserByEmailSql, email)

	user := &model.User{}
	err := row.Scan(&user.Name, &user.Password, &user.Online)
	if err != nil {
		fmt.Println("Error querying for user:", err.Error())
		return nil
	}

	return user
}

func (s *Sqlite) UpdatePlayer(player *model.Player) int64 {
	// TODO: stats fields
	stmt, err := s.db.Prepare(queries.UpdatePlayerSql)
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(
		player.Location.X,
		player.Location.Y,
		player.Stats.Xp,
		player.Name)
	checkErr(err)

	affected, err := res.RowsAffected()
	checkErr(err)

	// TODO: improve this lol
	// First delete all the items we don't have anymore
	knownItems := s.ReadItems(player.Name)
	for _, i := range knownItems {
		found := false
		for _, j := range player.Inventory {
			if j == nil {
				continue
			}

			if j.Id == i.Id {
				found = true
				break
			}
		}
		if found {
			continue
		}
		s.DeleteItem(i.Id)
	}

	// Then add the new items that aren't in the db
	for _, i := range player.Inventory {
		if i == nil {
			continue
		}

		found := false
		for _, j := range knownItems {
			if j.Id == i.Id {
				found = true
				break
			}
		}
		if found {
			continue
		}
		s.CreateItem(i)
	}

	return affected
}

func (s *Sqlite) updateUserStatus(name string, online int) error {
	stmt, err := s.db.Prepare(queries.UpdateUserSql)
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	res, err := stmt.Exec(online, name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (s *Sqlite) UpdateUserOnline(name string) error {
	return s.updateUserStatus(name, 1)
}

func (s *Sqlite) UpdateUserOffline(name string) error {
	return s.updateUserStatus(name, 0)
}

func (s *Sqlite) DeletePlayer(guid string) {
	// TODO
	// We don't delete, just mark 'enabled' to false/0
	return
}

func (s *Sqlite) CreateItem(item *model.Item) *model.Item {
	stmt, err := s.db.Prepare(queries.CreateItemSql)
	checkErr(err)
	defer stmt.Close()

	item.Id = uuid.New().String()

	res, err := stmt.Exec(
		item.Id,
		item.Name,
		item.Class,
		item.ItemLevel,
		item.Player)
	checkErr(err)

	_, err = res.RowsAffected()
	checkErr(err)

	return item
}

func (s *Sqlite) ReadItem(guid string) *model.Item {
	row := s.db.QueryRow(queries.ReadItemSql, guid)

	item := &model.Item{}
	err := row.Scan(&item.Id, &item.Name, &item.Class, &item.ItemLevel, &item.Player)
	if err != nil {
		return nil
	}

	return item
}

func (s *Sqlite) ReadItems(playerName string) []*model.Item {
	rows, err := s.db.Query(queries.ReadItemsByPlayerSql, playerName)
	defer rows.Close()

	checkErr(err)

	items := make([]*model.Item, 0)
	for rows.Next() {
		item := &model.Item{}

		err = rows.Scan(
			&item.Id,
			&item.Name,
			&item.Class,
			&item.ItemLevel,
			&item.Player)
		checkErr(err)

		items = append(items, item)
	}

	err = rows.Err()
	checkErr(err)

	return items
}

func (s *Sqlite) UpdateItem(*model.Item) int64 {
	// TODO
	return 0
}

func (s *Sqlite) DeleteItem(guid string) {
	stmt, err := s.db.Prepare(queries.DeleteItemSql)
	defer stmt.Close()

	checkErr(err)

	res, err := stmt.Exec(guid)
	checkErr(err)

	_, err = res.RowsAffected()
	checkErr(err)

	return
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
