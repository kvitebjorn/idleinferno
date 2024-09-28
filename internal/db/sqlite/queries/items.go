package queries

const CreateItemsTableSql string = `CREATE TABLE items (
	id	      TEXT PRIMARY KEY NOT NULL UNIQUE,
	name	    TEXT NOT NULL,
	xcoord    INTEGER,
	ycoord    INTEGER,
	class     INTEGER,
	itemlevel INTEGER,
	equipped  INTEGER,
	player    TEXT,
	FOREIGN KEY(player) REFERENCES players(id)
)`
