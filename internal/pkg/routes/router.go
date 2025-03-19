package routes

import (
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site"
)

func Router() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", site.Router())
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return router
}
