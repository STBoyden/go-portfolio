package migrations

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/lib/pq" // need the pq driver
)

// RunMigrations applies database schema migrations to a PostgreSQL database
// using the specified migration source. It loads environment variables from a
// predefined file and checks for the "DB_URL" variable (trimming any
// surrounding quotes). It then opens a PostgreSQL driver connection,
// initializes a migration instance with the given source, and executes the
// migration using the Up method, returning any errors encountered except for
// those indicating no changes.
func RunMigrations(dbURL, sourceName string, sourceInstance source.Driver) error {
	p := &migratePgx.Postgres{}
	driver, err := p.Open(dbURL)
	if err != nil {
		return err
	}
	defer driver.Close()

	migrations, err := migrate.NewWithInstance(sourceName, sourceInstance, "pgx", driver)
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
