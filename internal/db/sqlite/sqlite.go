package sqlite

import (
	"database/sql"
	"log"
	"os"

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
		log.Println("Creating database...")
		createTables = true
		_, err = os.Create(DatabaseName)
		if err != nil {
			log.Fatalln("Failed to create database: ", err.Error())
		}
		log.Println("Database created successfully!")
	}

	log.Println("Connecting to database...")
	db, err := sql.Open("sqlite3", DatabaseName)
	if err != nil {
		log.Fatalln("Error connecting to db: ", err)
	}
	log.Println("Database connection successful!")

	s.db = db

	if createTables {
		log.Println("Creating database tables...")
		_, err = db.Exec(queries.CreatePlayersTableSql)
		_, err = db.Exec(queries.CreateItemsTableSql)
		if err != nil {
			log.Fatalln("Failed to initialize database tables: ", err.Error())
		}
		log.Println("Database tables created successfully!")
	}

	return
}

func (s *Sqlite) Close() error {
	return s.db.Close()
}

func (s *Sqlite) CreatePlayer(*model.Player) *model.Player {
	// TODO
	return nil
}

func (s *Sqlite) ReadPlayer(guid string) *model.Player {
	// TODO
	return nil
}

func (s *Sqlite) ReadPlayers() []*model.Player {
	rows, err := s.db.Query(queries.ReadPlayersSql)
	defer rows.Close()

	checkErr(err)

	players := make([]*model.Player, 0)
	for rows.Next() {
		player := &model.Player{Location: &model.Coordinates{X: 0, Y: 0}}

		// TODO: scan the rest
		err = rows.Scan(&player.Id, &player.Name, &player.Location.X, &player.Location.Y)
		checkErr(err)

		players = append(players, player)
	}

	err = rows.Err()
	checkErr(err)

	return players
}

func (s *Sqlite) UpdatePlayer(player *model.Player) int64 {
	// TODO: rest of the fields
	stmt, err := s.db.Prepare(queries.UpdatePlayerSql)
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(player.Location.X, player.Location.Y, player.Id)
	checkErr(err)

	affected, err := res.RowsAffected()
	checkErr(err)

	return affected
}

func (s *Sqlite) DeletePlayer(guid string) {
	// TODO
	return
}

func (s *Sqlite) CreateItem(*model.Item) *model.Item {
	// TODO
	return nil
}

func (s *Sqlite) ReadItem(guid string) *model.Item {
	// TODO
	return nil
}

func (s *Sqlite) ReadItems() []*model.Item {
	// TODO
	return nil
}

func (s *Sqlite) UpdateItem(*model.Item) int64 {
	// TODO
	return 0
}

func (s *Sqlite) DeleteItem(guid string) {
	// TODO
	return
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}
