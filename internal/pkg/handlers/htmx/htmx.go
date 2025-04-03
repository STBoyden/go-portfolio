package htmx

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

func Handler(component templ.Component, options ...func(*templ.ComponentHandler)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render := component

		if _, ok := r.Header["Hx-Request"]; !ok {
			render = views.Root(component)
		}

		templ.Handler(render, options...).ServeHTTP(w, r)
	})
}
