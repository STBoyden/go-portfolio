//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
	"github.com/a-h/templ"
)

func Router() *http.ServeMux {
	root := views.Root

	router := http.NewServeMux()
	router.Handle("/", templ.Handler(root(views.Index())))

	return router
}
