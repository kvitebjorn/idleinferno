package queries

const CreatePlayersTableSql string = `CREATE TABLE players (
		id	      TEXT PRIMARY KEY NOT NULL UNIQUE,
		name	    TEXT NOT NULL UNIQUE,
		xcoord    INTEGER,
		ycoord    INTEGER,
		email	    TEXT NOT NULL UNIQUE,
		password	TEXT NOT NULL,
		class     TEXT NOT NULL,
		level     INTEGER,
		xp        INTEGER,
		itemlevel INTEGER,
		head      TEXT,
		torso     TEXT,
	 	legs      TEXT,
		arms      TEXT,
		gloves    TEXT,
		boots     TEXT,
		necklace  TEXT,
		ring      TEXT,
		weapon    TEXT,
		online    INTEGER,
		created   TEXT NOT NULL,
		enabled   INTEGER
	)`

// TODO: the rest of the fields
const (
	ReadPlayerSql   string = `SELECT id, name, xcoord, ycoord FROM players WHERE name = ?`
	ReadPlayersSql  string = `SELECT id, name, xcoord, ycoord FROM players`
	UpdatePlayerSql string = `UPDATE players SET xcoord = ?, ycoord = ? WHERE id = ?; `

	ReadUserSql   string = `SELECT name, password, online FROM players WHERE name = ?`
	UpdateUserSql string = `UPDATE players SET online = ? WHERE name = ?`
)
