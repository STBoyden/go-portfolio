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

	router.HandleFunc("GET /page/index", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Index().Render(r.Context(), w)
	})

	router.HandleFunc("GET /page/about", func(w http.ResponseWriter, r *http.Request) {
		_ = views.About().Render(r.Context(), w)
	})

	router.Handle("GET /blog", templ.Handler(views.Root(views.Blog())))
	router.HandleFunc("GET /blog/admin", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
		} else {
			request, err := http.NewRequest(http.MethodPost, "/api/v1/blog/check-authentication", nil)
			if err != nil {
				_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
				return
			}
			request.AddCookie(cookie)

			client := &http.Client{}

			response, err := client.Do(request)
			if err != nil {
				_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
				return
			}
			defer response.Body.Close()

			// showExpirationWarning := response.StatusCode == http.StatusAccepted
			if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusAccepted {
			} else {
				_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
				return
			}

			// _ = views.Root(views.BlogAdmin(showExpirationWarning)).Render(r.Context(), w)
		}
	})

	router.HandleFunc("GET /page/blog", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Blog().Render(r.Context(), w)
	})

	return router
}
