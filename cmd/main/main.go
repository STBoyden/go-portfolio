package main

import (
	"net/http"

	fs "github.com/STBoyden/go-portfolio"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes"
)

func main() {
	mux := http.NewServeMux()

	// Forward all endpoints to routes.Router()
	mux.Handle("/", routes.Router(fs.StaticFS))

	println("Serving http://localhost:8080...")
	http.ListenAndServe(":8080", mux)
}
