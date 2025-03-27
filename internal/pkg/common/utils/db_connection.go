package utils

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	*pgx.Conn
	context.Context
}

//nolint:gochecknoglobals // This is a global variable that is used to store the database connection.
var Database *DB

func ConnectDB() {
	if Database != nil {
		return
	}

	url := MustEnv("DB_URL")
	connectionContext := context.Background()

	db := Must(pgx.Connect(connectionContext, url))

	Database = &DB{Conn: db, Context: connectionContext}
}
