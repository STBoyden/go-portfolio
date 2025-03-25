package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	return value
}

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

	var url string
	var ok bool
	if url, ok = os.LookupEnv("DB_URL"); !ok {
		panic("DB_URL environment variable not set")
	}

	connectionContext := context.Background()

	db := must(pgx.Connect(connectionContext, url))

	Database = &DB{Conn: db, Context: connectionContext}
}
