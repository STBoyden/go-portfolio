//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/handlers/htmx"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

const siteLogTag string = "site"

func needAuthRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/new-post", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)
		log := w.LoggingMiddleware.Log

		if !w.Authed() {
			log(middleware.Info, siteLogTag, "user is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		htmx.Handler(views.BlogAdminPostEdit(nil, false), templ.WithStreaming()).ServeHTTP(w, r)
	})

	mux.HandleFunc("/edit/{id}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)
		log := w.LoggingMiddleware.Log

		if !w.Authed() {
			log(middleware.Info, siteLogTag, "user is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			log(middleware.Info, siteLogTag, "user did not provide valid uuid")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		queries := utils.Database.StartQueries()
		defer utils.Database.EndQueries()

		post, err := queries.GetPostByID(r.Context(), id)
		if err != nil {
			log(middleware.Info, siteLogTag, "ID '%v' did not exist in database", id)
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		htmx.Handler(views.BlogAdminPostEdit(&post, true), templ.WithStreaming()).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /preview/{slug}", func(_w http.ResponseWriter, r *http.Request) {
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

		queries := utils.Database.StartQueries()
		defer utils.Database.EndQueries()

		post, err := queries.GetPostBySlug(r.Context(), slug)
		if err != nil {
			log(middleware.Info, siteLogTag, "post with given slug could not be found: %v", err)
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		htmx.Handler(views.BlogPost(post, true)).ServeHTTP(w, r)
	})

	return mux
}

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", htmx.Handler(views.Index()))
	mux.Handle("GET /about", htmx.Handler(views.About()))
	mux.Handle("GET /blog", htmx.Handler(views.Blog()))

	mux.HandleFunc("GET /blog/post/{slug}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		slug := r.PathValue("slug")
		if slug == "" {
			w.Log(middleware.Info, "http", "slug path value was invalid")
			http.NotFound(w, r)
			return
		}

		queries := utils.Database.StartQueries()
		defer utils.Database.EndQueries()

		post, err := queries.GetPublishedPostBySlug(r.Context(), slug)
		if err != nil {
			w.Log(middleware.Info, "http", "could not get post: %v", err)
			http.NotFound(w, r)
			return
		}

		htmx.Handler(views.BlogPost(post, false)).ServeHTTP(w, r)
	})

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

		htmx.Handler(views.BlogAdminAuthed(showExpirationWarning)).ServeHTTP(w, r)
	})

	mux.Handle("/blog/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/blog/admin", needAuthRoutes())))

	return mux
}
