package main

import (
	"database/sql"
	"fmt"
	"strings"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/gotenv/v2"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	env, _ := gotenv.LoadEnvFromFS(fs.EnvFile)

	var dbUrl string
	if s, ok := env["DB_URL"]; !ok {
		panic("DB_URL environment variable not defined")
	} else {
		dbUrl = strings.Trim(s, "\"")
	}

	db := utils.Must(sql.Open("postgres", dbUrl))
	driver := utils.Must(pgx.WithInstance(db, &pgx.Config{}))
	migrations := utils.Must(migrate.NewWithDatabaseInstance("file://./migrations/", "postgres", driver))

	err := migrations.Up()
	if err != nil && !strings.Contains(err.Error(), "no change") {
		panic(fmt.Sprintf("migrations didn't run successfully: %v", err))
	} else {
		fmt.Println("migrations ran successfully or no change")
	}
}
