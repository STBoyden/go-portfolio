package v1

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/types"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
)

//nolint:gochecknoglobals // These are only accessible in the v1 package, and are not globally accessible by other packages.
var (
	adminUsername string
	adminPassword string
)

const blogAuthLogTag string = "blog-auth"

func blogAdmin() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /new-post/{slug}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if _, authed := w.Details(); !authed {
			w.Log(middleware.Info, "user is not authorised to create a new post")
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		reader, err := r.GetBody()
		if err != nil {
			w.Log(middleware.Info, "given request has no body. body len <= 0")
			w.PrepareHeader(http.StatusBadRequest)
			return
		}
		defer reader.Close()

		buffer, err := io.ReadAll(reader)
		if err != nil {
			w.Log(middleware.Info, "body was malformed and could not be read properly")
			w.PrepareHeader(http.StatusBadRequest)
			return
		}

		slug := r.PathValue("slug")
		if slug == "" {
			w.Log(middleware.Info, "slug was not present in path")
			w.PrepareHeader(http.StatusBadRequest)
			return
		}

		blogContent := types.BlogContent{}
		err = json.Unmarshal(buffer, &blogContent)
		if err != nil {
			w.Log(middleware.Info, "body was not in the correct format and could not be parsed: %v", err)
			w.PrepareHeader(http.StatusBadRequest)
			return
		}

		queries := persistence.New(utils.Database)
		post, err := queries.CreatePost(r.Context(), persistence.CreatePostParams{Slug: slug, Content: buffer})
		if err != nil {
			panic(fmt.Sprintf("was unable to insert a new post: %v", err))
		}

		type response struct {
			PostID uuid.UUID `json:"post_id"`
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response{PostID: post.ID})
		if err != nil {
			panic(fmt.Sprintf("unable to create response object: %v", err))
		}
	})

	mux.HandleFunc("GET /posts", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if _, authed := w.Details(); !authed {
			w.Log(middleware.Info, "user is not authorised to get unpublished posts")
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		queries := persistence.New(utils.Database)
		posts, err := queries.GetAllPosts(r.Context())
		if err != nil {
			_ = components.Error().Render(r.Context(), w)
			return
		}

		templ.Handler(components.PostList(posts, true), templ.WithStreaming()).ServeHTTP(w, r)
	})

	return mux
}

// checkAuthentication checks the authentication of the request and responds
// with whether the request has valid authentication.
func checkAuthentication(_w http.ResponseWriter, r *http.Request) {
	w := utils.MustCast[middleware.LoggingMiddleware](_w)

	cookie, err := r.Cookie(consts.TokenCookieName)
	if err != nil {
		w.Log(middleware.Info, blogAuthLogTag, "token cookie was missing from client request")
		w.PrepareHeader(http.StatusUnauthorized)
		return
	}

	token, err := uuid.Parse(cookie.Value)
	if err != nil {
		w.Log(middleware.Info, blogAuthLogTag, "token form cookie was invalid")
		w.PrepareHeader(http.StatusUnauthorized)
		return
	}

	queries := persistence.New(utils.Database)
	authorisation, err := queries.GetAuthByToken(r.Context(), token)
	if err != nil {
		w.Log(middleware.Warn, blogAuthLogTag, "internal error occurred: de-authing user just in case: could not get authorisation token: %v", err)
		w.PrepareHeader(http.StatusUnauthorized)
		return
	}

	if authorisation.Expiry.Before(time.Now()) {
		w.Log(middleware.Info, blogAuthLogTag, "token associated with request has expired")
		w.PrepareHeader(http.StatusUnauthorized)
		return
	} else if authorisation.Expiry.Before(time.Now().Add(30 * time.Minute)) {
		// if the token is within 30 minutes of expiration, return a 202
		// status code so that the front-end may warn the user.
		w.PrepareHeader(http.StatusAccepted)
	}
}

// authenticate creates a new authentication for a user if they have provided
// the correct login details and returns a cookie with a auth token.
func authenticate(_w http.ResponseWriter, r *http.Request) {
	w := utils.MustCast[middleware.LoggingMiddleware](_w)

	onError := func(statusCode int) {
		w.PrepareHeader(statusCode)
		_ = components.Error().Render(r.Context(), w)
	}

	headerContent, ok := r.Header["Authorization"]
	if !ok {
		w.Log(middleware.Info, blogAuthLogTag, "authorization header missing")
		onError(http.StatusBadRequest)
		return
	}

	authorisation := strings.Join(headerContent, " ")
	if authorisation == "" {
		w.Log(middleware.Info, blogAuthLogTag, "authorization header content is empty")
		onError(http.StatusBadRequest)
		return
	}

	username, password, ok := r.BasicAuth()
	if username == "" || password == "" || !ok {
		w.Log(middleware.Info, blogAuthLogTag, "given username and/or password are empty")
		onError(http.StatusBadRequest)
		return
	}

	h := sha512.Sum512([]byte(password))
	passwordHashed := hex.EncodeToString(h[:])

	if username != adminUsername || passwordHashed != adminPassword {
		w.Log(middleware.Info, blogAuthLogTag, "given username and/or password hash does not match administrator details")
		onError(http.StatusUnauthorized)
		return
	}

	queries := persistence.New(utils.Database)
	auth, err := queries.NewAuth(r.Context())
	if err != nil {
		panic(fmt.Sprintf("unable to create a new token: %v", err))
	}

	w.Header().Add("Hx-Trigger", "login-page-reload")
	http.SetCookie(w, &http.Cookie{
		Name:     consts.TokenCookieName,
		Value:    auth.ID.String(),
		Expires:  auth.Expiry,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Log(middleware.Info, blogAuthLogTag, "setting cookie %s", consts.TokenCookieName)
}

func BlogAPI() *http.ServeMux {
	adminUsername = utils.MustEnv("ADMIN_USER")
	adminPassword = utils.MustEnv("ADMIN_PW")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /posts", func(w http.ResponseWriter, r *http.Request) {
		queries := persistence.New(utils.Database)
		posts, err := queries.GetPublishedPosts(r.Context())
		if err != nil {
			_ = components.Error().Render(r.Context(), w)
			return
		}

		templ.Handler(components.PostList(posts, false), templ.WithStreaming()).ServeHTTP(w, r)
	})

	mux.Handle("/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/admin", blogAdmin())))

	// responds with a status code relevant to the authentication status of the user.
	// 200: the user is authenticated and cookie is outside of a 30 minute
	// 		expiration warning.
	// 202: the user is authenticated and cookie is within a 30 minute
	// 		expiration warning.
	// 401: the user is not authenticated.
	mux.HandleFunc("POST /check-authentication", checkAuthentication)

	// authenticates a user to be able to use the /admin/ endpoints and redirects to the page
	mux.HandleFunc("POST /authenticate", authenticate)

	return mux
}
