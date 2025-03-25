package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes"
	"github.com/STBoyden/gotenv/v2"
)

const (
	ReadHeaderTimeout = 5 * time.Second
	WriteTimeout      = 10 * time.Second
	IdleTimeout       = 15 * time.Second
)

func main() {
	_, _ = gotenv.LoadEnvFromFS(fs.EnvFile)
	utils.ConnectDB()
	defer utils.Database.Close(utils.Database.Context)

	mux := http.NewServeMux()

	// Forward all endpoints to routes.Router()
	mux.Handle("/", middleware.Handlers.Logger(routes.Router(fs.StaticFS)))

	log.Println("Serving http://localhost:8080...")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: ReadHeaderTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("could not listen and serve to :8080: %v", err))
	}
}
