package db

import (
	"database/sql"
	"log"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	db "github.com/spiritov/svelte-go-steam/db/queries"
)

var (
	Queries *db.Queries
)

func OpenDB(path string) *sql.DB {
	// os.Remove("./db/jump.db")
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.Error("failed to open db", "error", err)
		log.Fatal()
	}

	Queries = db.New(database)
	return database
}
