package migrations

import (
	"errors"
	"log"
	"strings"

	"github.com/STBoyden/gotenv/v2"
	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/lib/pq" // need the pq driver

	fs "github.com/STBoyden/go-portfolio"
)

var ErrDatabaseEnvironmentVariableNotSet = errors.New("DB_URL environment variable not set")

// RunMigrations applies database schema migrations to a PostgreSQL database
// using the specified migration source. It loads environment variables from a
// predefined file and checks for the "DB_URL" variable (trimming any
// surrounding quotes). If "DB_URL" is missing, it returns
// [ErrDatabaseEnvironmentVariableNotSet]. It then opens a PostgreSQL driver
// connection, initializes a migration instance with the given source, and
// executes the migration using the Up method, returning any errors encountered
// except for those indicating no changes.
func RunMigrations(source string) error {
	env, _ := gotenv.LoadEnvFromFS(fs.EnvFile, gotenv.LoadOptions{OverrideExistingVars: false})

	var dbURL string
	var ok bool
	if dbURL, ok = env["DB_URL"]; !ok {
		return ErrDatabaseEnvironmentVariableNotSet
	}

	dbURL = strings.Trim(dbURL, "\"")

	p := &migratePgx.Postgres{}
	driver, err := p.Open(dbURL)
	if err != nil {
		return err
	}
	defer driver.Close()

	migrations, err := migrate.NewWithDatabaseInstance(source, "pgx", driver)
	if err != nil {
		return err
	}

	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Print("Migrations ran successfully or no change")

	return nil
}
