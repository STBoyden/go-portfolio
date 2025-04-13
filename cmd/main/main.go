package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/STBoyden/gotenv/v2"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/migrations"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes"
	"github.com/STBoyden/go-portfolio/internal/pkg/utils"
)

const (
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 15 * time.Second
)

func main() {
	_, _ = gotenv.LoadEnvFromFS(fs.EnvFile, gotenv.LoadOptions{OverrideExistingVars: false})

	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	w := os.Stderr
	logger := slog.New(tint.NewHandler(colorable.NewColorable(w), &tint.Options{
		TimeFormat: time.DateTime + " MST (-0700)",
		Level:      level,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}))
	slog.SetDefault(logger)

	dbURL := utils.MustEnv("DB_URL")

	d, err := iofs.New(fs.MigrationsFS, "migrations")
	if err != nil {
		logger.Error("Could not get migrations", "err", err)
		return
	}

	err = migrations.RunMigrations(dbURL, "iofs", d)
	if err != nil {
		logger.Error("Could not run migrations on database", "err", err)
		return
	}

	utils.ConnectDB()
	defer utils.Database.Close(context.Background())

	mux := http.NewServeMux()

	// Forward all endpoints to routes.Router()
	mux.Handle("/", routes.Router(fs.StaticFS))

	logger.Info("Serving http://localhost:8080...")
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	err = server.ListenAndServe()
	if err != nil {
		logger.Error("Could not listen and serve to :8080", "err", err)
		return
	}
}
