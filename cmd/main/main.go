package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/STBoyden/gotenv/v2"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/migrations"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes"
)

const (
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 15 * time.Second
)

func main() {
	_, _ = gotenv.LoadEnvFromFS(fs.EnvFile)

	err := migrations.RunMigrations("file://./migrations/")
	if err != nil {
		panic("couldn't run migrations on database")
	}

	utils.ConnectDB()
	defer utils.Database.Close(utils.Database.Context)

	mux := http.NewServeMux()

	// Forward all endpoints to routes.Router()
	mux.Handle("/", routes.Router(fs.StaticFS))

	log.Println("Serving http://localhost:8080...")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("could not listen and serve to :8080: %v", err))
	}
}
