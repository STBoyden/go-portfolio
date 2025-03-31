package routes

import (
	"embed"
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	v1 "github.com/STBoyden/go-portfolio/internal/pkg/routes/api/v1"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site"
)

func Router(static embed.FS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", middleware.Handlers.Logger(site.Router()))
	mux.Handle("/api/v1/", middleware.Handlers.Logger(http.StripPrefix("/api/v1", v1.Router())))
	mux.Handle("/static/", middleware.Handlers.Logger(http.FileServer(http.FS(static))))

	return mux
}
