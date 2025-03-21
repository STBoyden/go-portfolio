package routes

import (
	"embed"
	"net/http"

	v1 "github.com/STBoyden/go-portfolio/internal/pkg/routes/api/v1"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site"
)

func Router(static embed.FS) *http.ServeMux {
	r := http.NewServeMux()
	r.Handle("/", site.Router())
	r.Handle("/api/v1/", http.StripPrefix("/api/v1", v1.Router()))
	r.Handle("/static/", http.FileServer(http.FS(static)))

	return r
}
