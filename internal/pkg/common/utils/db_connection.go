package utils

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
)

// ReadOnlyQueries contains queries that only read from the database.
type ReadOnlyQueries interface {
	CheckAuthExists(ctx context.Context, id uuid.UUID) (bool, error)
	CheckIfAuthExpired(ctx context.Context, id uuid.UUID) (bool, error)
	GetAuthByToken(ctx context.Context, id uuid.UUID) (persistence.Authorisation, error)
	GetExpiredAuths(ctx context.Context) ([]persistence.Authorisation, error)

	GetAllPosts(ctx context.Context) ([]persistence.Post, error)
	GetPostByID(ctx context.Context, id uuid.UUID) (persistence.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (persistence.Post, error)
	GetPublishedPostBySlug(ctx context.Context, slug string) (persistence.Post, error)
	GetPublishedPosts(ctx context.Context) ([]persistence.Post, error)
}

var _ ReadOnlyQueries = (*persistence.Queries)(nil)

// AppDatabase.
type AppDatabase struct {
	conn    *pgx.Conn
	queries *persistence.Queries
}

type (
	CommitFunc   func(context.Context) error
	RollbackFunc func(context.Context)
)

func (a *AppDatabase) rollbackBuilder(tx pgx.Tx) RollbackFunc {
	return func(ctx context.Context) {
		_ = tx.Rollback(ctx)
	}
}

func (a *AppDatabase) commitBuilder(tx pgx.Tx) CommitFunc {
	return func(ctx context.Context) error {
		return tx.Commit(ctx)
	}
}

// StartWriteTx starts a new transaction and locks writes and reads until
// done. Nested transactions are not supported.
//
// This function returns four values: a [persistence.Queries] value that has
// been wrapped in a transaction, a [CommitFunc] and [RollbackFunc] that are
// used to commit and rollback changes accordingly, and finally an error value
// that if not nil, means all other values *are* nil.
func (a *AppDatabase) StartWriteTx(ctx context.Context) (*persistence.Queries, CommitFunc, RollbackFunc, error) {
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	return a.queries.WithTx(tx), a.commitBuilder(tx), a.rollbackBuilder(tx), nil
}

// StartReadTx starts a read from the database. This function starts a read
// lock on the database, and multiple goroutines can call this function at the
// same time without blocking, however this function will block if there's an
// active write on the database from [NewTransaction].
func (a *AppDatabase) StartReadTx(ctx context.Context) (ReadOnlyQueries, CommitFunc, RollbackFunc, error) {
	tx, err := a.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	return a.queries.WithTx(tx), a.commitBuilder(tx), a.rollbackBuilder(tx), nil
}

// Close closes the connection to the database. See [pgx.Conn.Close] for more
// information.
func (a *AppDatabase) Close(ctx context.Context) error {
	return a.conn.Close(ctx)
}

//nolint:gochecknoglobals // This is a global variable that is used to store the database connection.
var Database *AppDatabase

func ConnectDB() {
	if Database != nil {
		return
	}

	url := MustEnv("DB_URL")
	connectionContext := context.Background()

	db := Must(pgx.Connect(connectionContext, url))
	queries := persistence.New(db)

	Database = &AppDatabase{conn: db, queries: queries}
}
