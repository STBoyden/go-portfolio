//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

const siteLogTag string = "site"

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", templ.Handler(views.Root(views.Index())))
	mux.Handle("GET /about", templ.Handler(views.Root(views.About())))

	mux.HandleFunc("GET /page/index", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Index().Render(r.Context(), w)
	})

	mux.HandleFunc("GET /page/about", func(w http.ResponseWriter, r *http.Request) {
		_ = views.About().Render(r.Context(), w)
	})

	mux.Handle("GET /blog", templ.Handler(views.Root(views.Blog())))
	mux.HandleFunc("GET /blog/admin/{$}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		cookie, err := r.Cookie(consts.TokenCookieName)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "token cookie missing from request")
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
			return
		} else if err = cookie.Valid(); err != nil {
			w.Log(middleware.Info, siteLogTag, "cookie is invalid: %v", err)
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
			return
		}

		request, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/api/v1/blog/check-authentication", nil)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "could not check that request was authenticated properly: %v", err)
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
			return
		}
		request.AddCookie(cookie)

		client := &http.Client{}

		response, err := client.Do(request)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "could not check that request was authenticated properly: %v", err)
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
			return
		}
		defer response.Body.Close()

		showExpirationWarning := response.StatusCode == http.StatusAccepted
		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
			w.Log(middleware.Info, siteLogTag, "token is not authorised")
			_ = views.Root(views.BlogAdminLogin()).Render(r.Context(), w)
			return
		}

		_ = views.Root(views.BlogAdminAuthed(showExpirationWarning)).Render(r.Context(), w)
	})

	mux.Handle("GET /blog/admin/new-post", middleware.Handlers.Authorisation(http.HandlerFunc(func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)
		log := w.LoggingMiddleware.Log

		if !w.Authed() {
			log(middleware.Info, siteLogTag, "user is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		id, _ := uuid.NewRandom()
		templ.Handler(views.BlogAdminPostEdit(&id), templ.WithStreaming()).ServeHTTP(w, r)
	})))

	mux.HandleFunc("GET /page/blog", func(w http.ResponseWriter, r *http.Request) {
		_ = views.Blog().Render(r.Context(), w)
	})

	return mux
}
