package main

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/STBoyden/go-portfolio/internal/pkg/migrations"
)

// main is the entry point for running database migrations. It applies migrations from the "./migrations/" directory and panics if an error occurs during the migration process.
func main() {
	err := migrations.RunMigrations("file://./migrations/")
	if err != nil {
		panic("could not run migrations: " + err.Error())
	}
}
