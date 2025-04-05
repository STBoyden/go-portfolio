//go:generate go tool github.com/a-h/templ/cmd/templ generate views
package site

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/handlers/htmx"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views"
)

func siteLog(logger *middleware.LoggingMiddleware, level middleware.LoggerLevel, message string, v ...any) {
	logger.Log(level, message, slog.Group("site", v...))
}

func needAuthRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/new-post", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if !w.Authed() {
			siteLog(w.LoggingMiddleware, middleware.Info, "User is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		htmx.Handler(views.BlogAdminPostEdit(nil, false), templ.WithStreaming()).ServeHTTP(w, r)
	})

	mux.HandleFunc("/edit/{id}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if !w.Authed() {
			siteLog(w.LoggingMiddleware, middleware.Info, "User is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			siteLog(w.LoggingMiddleware, middleware.Info, "User did not provide valid uuid")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		queries, commit, rollback, err := utils.Database.StartReadTx(r.Context())
		if err != nil {
			panic("unable to start transaction on database: " + err.Error())
		}
		defer rollback(r.Context())

		post, err := queries.GetPostByID(r.Context(), id)
		if err != nil {
			siteLog(w.LoggingMiddleware, middleware.Info, "ID did not exist in database", "id", id)
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		_ = commit(r.Context())

		htmx.Handler(views.BlogAdminPostEdit(&post, true), templ.WithStreaming()).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /preview/{slug}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if !w.Authed() {
			siteLog(w.LoggingMiddleware, middleware.Info, "User is not authorised")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		slug := r.PathValue("slug")
		if slug == "" {
			siteLog(w.LoggingMiddleware, middleware.Info, "Slug not provided")
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		queries, commit, rollback, err := utils.Database.StartReadTx(r.Context())
		if err != nil {
			panic("unable to start transaction on database: " + err.Error())
		}
		defer rollback(r.Context())

		post, err := queries.GetPostBySlug(r.Context(), slug)
		if err != nil {
			siteLog(w.LoggingMiddleware, middleware.Info, "Post with given slug could not be found", "error", err)
			http.Redirect(w, r, "/blog/admin", http.StatusSeeOther)
			return
		}

		_ = commit(r.Context())

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
			w.Log(middleware.Info, "Slug path value was invalid")
			http.NotFound(w, r)
			return
		}

		queries, commit, rollback, err := utils.Database.StartReadTx(r.Context())
		if err != nil {
			panic("unable to start transaction on database: " + err.Error())
		}
		defer rollback(r.Context())

		post, err := queries.GetPublishedPostBySlug(r.Context(), slug)
		if err != nil {
			w.Log(middleware.Info, "Could not get post", "error", err)
			http.NotFound(w, r)
			return
		}

		_ = commit(r.Context())

		htmx.Handler(views.BlogPost(post, false)).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /blog/admin", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		component := views.Root(views.BlogAdminLogin())

		cookie, err := r.Cookie(consts.TokenCookieName)
		if err != nil {
			siteLog(w, middleware.Info, "Token cookie missing from request")
			templ.Handler(component).ServeHTTP(w, r)
			return
		} else if err = cookie.Valid(); err != nil {
			siteLog(w, middleware.Info, "Cookie is invalid", "error", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}

		request, err := http.NewRequest(http.MethodPost, "http://0.0.0.0:8080/api/v1/blog/check-authentication", nil)
		if err != nil {
			siteLog(w, middleware.Info, "Could not check that request was authenticated properly", "error", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		request.AddCookie(cookie)

		client := &http.Client{}

		response, err := client.Do(request)
		if err != nil {
			siteLog(w, middleware.Info, "Could not check that request was authenticated properly", "error", err)
			templ.Handler(component).ServeHTTP(w, r)
			return
		}
		defer response.Body.Close()

		showExpirationWarning := response.StatusCode == http.StatusAccepted
		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
			siteLog(w, middleware.Info, "Token is not authorised")
			templ.Handler(component).ServeHTTP(w, r)
			return
		}

		htmx.Handler(views.BlogAdminAuthed(showExpirationWarning)).ServeHTTP(w, r)
	})

	mux.Handle("/blog/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/blog/admin", needAuthRoutes())))

	return mux
}
