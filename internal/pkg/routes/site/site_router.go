//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

const siteLogTag string = "site"

func needAuthRoutes() http.Handler {
	submux := http.NewServeMux()

	submux.HandleFunc("/new-post", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)
		log := w.LoggingMiddleware.Log

		if !w.Authed() {
			log(middleware.Info, siteLogTag, "user is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		templ.Handler(views.BlogAdminPostEdit(nil, false), templ.WithStreaming()).ServeHTTP(w, r)
	})

	submux.HandleFunc("GET /preview/{slug}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)
		log := w.LoggingMiddleware.Log

		if !w.Authed() {
			log(middleware.Info, siteLogTag, "user is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		slug := r.PathValue("slug")
		if slug == "" {
			log(middleware.Info, siteLogTag, "slug not provided")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		queries := persistence.New(utils.Database)
		post, err := queries.GetPostBySlug(r.Context(), slug)
		if err != nil {
			log(middleware.Info, siteLogTag, "post with given slug could not be found: %v", err)
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		templ.Handler(views.Root(views.BlogPost(post, true))).ServeHTTP(w, r)
	})

	return submux
}

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", templ.Handler(views.Root(views.Index())))
	mux.Handle("GET /about", templ.Handler(views.Root(views.About())))
	mux.Handle("GET /page/index", templ.Handler(views.Index()))
	mux.Handle("GET /page/about", templ.Handler(views.About()))
	mux.Handle("GET /blog", templ.Handler(views.Root(views.Blog())))

	mux.HandleFunc("GET /blog/admin", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		component := views.Root(views.BlogAdminLogin())

		cookie, err := r.Cookie(consts.TokenCookieName)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "token cookie missing from request")
			templ.Handler(component).ServeHTTP(w, r)
			return
		} else if err = cookie.Valid(); err != nil {
			w.Log(middleware.Info, siteLogTag, "cookie is invalid: %v", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}

		request, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/api/v1/blog/check-authentication", nil)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "could not check that request was authenticated properly: %v", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		request.AddCookie(cookie)

		client := &http.Client{}

		response, err := client.Do(request)
		if err != nil {
			w.Log(middleware.Info, siteLogTag, "could not check that request was authenticated properly: %v", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		defer response.Body.Close()

		showExpirationWarning := response.StatusCode == http.StatusAccepted
		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
			w.Log(middleware.Info, siteLogTag, "token is not authorised")
			templ.Handler(component).ServeHTTP(w, r)
			return
		}

		component = views.Root(views.BlogAdminAuthed(showExpirationWarning))
		templ.Handler(component).ServeHTTP(w, r)
	})

	mux.Handle("/blog/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/blog/admin", needAuthRoutes())))
	mux.Handle("GET /page/blog/{$}", templ.Handler(views.Blog()))

	return mux
}
