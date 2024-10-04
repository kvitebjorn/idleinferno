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
		online    INTEGER,
		created   TEXT NOT NULL,
		enabled   INTEGER
	)`

const (
	CreatePlayerSql string = `INSERT INTO players
	(id, name, email, password, class, xcoord, ycoord, level, xp, online, created, enabled)
	VALUES (?, ?, ?, ?, ?, 0, 0, 1, 0, 0, datetime(), 1)`
	ReadPlayerSql   string = `SELECT id, name, class, xcoord, ycoord, xp, level, created, online FROM players WHERE name = ?`
	ReadPlayersSql  string = `SELECT id, name, class, xcoord, ycoord, xp, level, created, online FROM players`
	UpdatePlayerSql string = `UPDATE players SET xcoord = ?, ycoord = ? WHERE name = ?;`

	ReadUserSql        string = `SELECT name, password, online FROM players WHERE name = ?`
	ReadUserByEmailSql string = `SELECT name, password, online FROM players WHERE email = ?`
	UpdateUserSql      string = `UPDATE players SET online = ? WHERE name = ?`
)
