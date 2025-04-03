// htmx contains a Handler function that handles the rendering of pages
// depending on received HTMX attributes.
//
// See: [Handler]
package htmx

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

// Simple handler to optionally render views.Root alongside the provided
// component depending on the HX-Request header. This provides the benefit of
// being able to get a partial and a full page from the same endpoint, with the
// decision being made on the presence of the header - allowing a cleaner SPA
// feel.
func Handler(component templ.Component, options ...func(*templ.ComponentHandler)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render := component

		if _, ok := r.Header["Hx-Request"]; !ok {
			render = views.Root(component)
		}

		templ.Handler(render, options...).ServeHTTP(w, r)
	})
}
