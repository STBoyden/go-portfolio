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

var Database *DB = nil

func ConnectDB() {
	if Database != nil {
		return
	}

	var url string
	if s, ok := os.LookupEnv("DB_URL"); !ok {
		panic("DB_URL environment variable not set")
	} else {
		url = s
	}

	connectionContext := context.Background()

	db := must(pgx.Connect(connectionContext, url))

	Database = &DB{Conn: db, Context: connectionContext}
}
