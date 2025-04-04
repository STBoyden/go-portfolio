package main

import (
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/STBoyden/gotenv/v2"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/migrations"
)

func main() {
	_, _ = gotenv.LoadEnvFromFS(fs.EnvFile, gotenv.LoadOptions{OverrideExistingVars: false})
	dbURL := utils.MustEnv("DB_URL")

	d, err := iofs.New(fs.MigrationsFS, "migrations")
	if err != nil {
		panic("could not get migrations:" + err.Error())
	}

	err = migrations.RunMigrations(dbURL, "iofs", d)
	if err != nil {
		panic("could not run migrations: " + err.Error())
	}
}
