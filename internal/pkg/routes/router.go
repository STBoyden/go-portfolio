package routes

import (
	"embed"
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site"
)

func Router(static embed.FS) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", site.Router())
	router.Handle("/static/", http.FileServer(http.FS(static)))

	return router
}
