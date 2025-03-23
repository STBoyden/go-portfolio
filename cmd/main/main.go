package main

import (
	"fmt"
	"log"
	"net/http"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes"
	"github.com/STBoyden/gotenv/v2"
)

func main() {
	_, _ = gotenv.LoadEnvFromFS(fs.EnvFile)
	utils.ConnectDB()
	defer utils.Database.Close(utils.Database.Context)

	mux := http.NewServeMux()

	// Forward all endpoints to routes.Router()
	mux.Handle("/", middleware.Logging(routes.Router(fs.StaticFS)))

	log.Println("Serving http://localhost:8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(fmt.Sprintf("could not listen and serve to :8080: %v", err))
	}
}
