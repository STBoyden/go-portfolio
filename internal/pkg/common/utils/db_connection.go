package utils

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"

	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
)

type appDatabase struct {
	mu sync.Mutex
	*pgx.Conn
	context.Context

	querying bool
	queries  *persistence.Queries
}

func (a *appDatabase) NewTransaction(ctx context.Context) (*persistence.Queries, pgx.Tx, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	tx, err := a.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	return a.queries.WithTx(tx), tx, nil
}

func (a *appDatabase) StartQueries() *persistence.Queries {
	a.mu.Lock()
	a.querying = true
	return a.queries
}

func (a *appDatabase) EndQueries() {
	if a.querying {
		a.querying = false
		a.mu.Unlock()
	}
}

//nolint:gochecknoglobals // This is a global variable that is used to store the database connection.
var Database *appDatabase

func ConnectDB() {
	if Database != nil {
		return
	}

	url := MustEnv("DB_URL")
	connectionContext := context.Background()

	db := Must(pgx.Connect(connectionContext, url))
	queries := persistence.New(db)

	Database = &appDatabase{Conn: db, queries: queries, Context: connectionContext}
}
