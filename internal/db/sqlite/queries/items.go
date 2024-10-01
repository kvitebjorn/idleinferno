package queries

const CreateItemsTableSql string = `CREATE TABLE items (
	id	      TEXT PRIMARY KEY NOT NULL UNIQUE,
	name	    TEXT NOT NULL,
	class     INTEGER,
	itemlevel INTEGER,
	player    TEXT,
	FOREIGN KEY(player) REFERENCES players(name)
)`
