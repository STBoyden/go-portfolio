package main

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/STBoyden/go-portfolio/internal/pkg/migrations"
)

func main() {
	err := migrations.RunMigrations("file://./migrations/")
	if err != nil {
		panic("could not run migrations: " + err.Error())
	}
}
