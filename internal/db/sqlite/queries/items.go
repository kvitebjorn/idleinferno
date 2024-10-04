package queries

const CreateItemsTableSql string = `CREATE TABLE items (
	id	      TEXT PRIMARY KEY NOT NULL UNIQUE,
	name	    TEXT NOT NULL,
	class     INTEGER,
	itemlevel INTEGER,
	player    TEXT,
	FOREIGN KEY(player) REFERENCES players(name)
)`

const (
	CreateItemSql        string = `INSERT INTO items (id, name, class, itemLevel, player) VALUES (?, ?, ?, ?, ?)`
	ReadItemSql          string = `SELECT id, name, class, itemLevel, player FROM items WHERE id = ?`
	ReadItemsByPlayerSql string = `SELECT id, name, class, itemLevel, player FROM items WHERE player = ?`
	DeleteItemSql        string = `DELETE FROM items WHERE id = ?`
)
