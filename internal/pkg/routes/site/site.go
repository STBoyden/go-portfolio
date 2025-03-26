//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

func Router() *http.ServeMux {
	router := http.NewServeMux()

	router.Handle("GET /{$}", templ.Handler(views.Root(views.Index())))
	router.Handle("GET /about", templ.Handler(views.Root(views.About())))
	router.Handle("GET /blog", templ.Handler(views.Root(views.Blog())))

	router.HandleFunc("GET /page/index", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Index().Render(r.Context(), w)
	})

	router.HandleFunc("GET /page/about", func(w http.ResponseWriter, r *http.Request) {
		_ = views.About().Render(r.Context(), w)
	})

	router.HandleFunc("GET /page/blog", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Blog().Render(r.Context(), w)
	})

	return router
}
