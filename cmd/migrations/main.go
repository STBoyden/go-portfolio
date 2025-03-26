package main

import (
	"fmt"
	"strings"

	"github.com/STBoyden/gotenv/v2"
	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
)

func main() {
	env, _ := gotenv.LoadEnvFromFS(fs.EnvFile, gotenv.LoadOptions{OverrideExistingVars: false})

	var dbURL string
	var ok bool
	if dbURL, ok = env["DB_URL"]; !ok {
		panic("DB_URL environment variable not defined")
	}

	dbURL = strings.Trim(dbURL, "\"")

	p := &migratePgx.Postgres{}
	driver := utils.Must(p.Open(dbURL))
	migrations := utils.Must(migrate.NewWithDatabaseInstance("file://./migrations/", "pgx", driver))

	err := migrations.Up()
	if err != nil && !strings.Contains(err.Error(), "no change") {
		panic(fmt.Sprintf("migrations didn't run successfully: %v", err))
	}

	fmt.Println("migrations ran successfully or no change")
}
