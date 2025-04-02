package v1

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

var (
	errMissingAuthorization        = errors.New("authorization header is missing")
	errInvalidAuthorizationContent = errors.New("invalid authorization content")
	errIncorrectCredentials        = errors.New("incorrect credentials")
)

const blogAuthLogTag string = "blog-auth"

func blogAdmin() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /new-post", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if err := r.ParseForm(); err != nil {
			w.Log(middleware.Info, "form not submitted correctly: %v", err)
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		type form struct {
			Title   string `json:"title"`
			Slug    string `json:"slug"`
			Content string `json:"content"`
		}

		if _, authed := w.Details(); !authed {
			w.Log(middleware.Info, "user is not authorised to create a new post")
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		title := r.PostFormValue("title")
		slug := r.PostFormValue("slug")
		content := r.PostFormValue("content")
		if title == "" || slug == "" || content == "" {
			w.Log(middleware.Info, "form content is invalid")
		}

		blogContent := types.BlogContent{
			Title: title,
			Text:  content,
		}
		contentBuffer, err := json.Marshal(blogContent)
		if err != nil {
			panic(fmt.Sprintf("unable to marshal blog content: %v", err))
		}

		queries := persistence.New(utils.Database)
		post, err := queries.CreatePost(r.Context(), persistence.CreatePostParams{
			Slug:    slug,
			Content: contentBuffer,
		})
		if err != nil {
			panic(fmt.Sprintf("was unable to insert a new post: %v", err))
		}

		http.Redirect(w, r, fmt.Sprintf("/blog/%s", post.Slug), http.StatusFound)
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

		component := components.PostList(posts, true)
		if err != nil {
			component = components.Error(err)
			return
		}

		templ.Handler(component, templ.WithStreaming()).ServeHTTP(w, r)
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

	onError := func(err error, statusCode int) {
		w.PrepareHeader(statusCode)
		templ.Handler(components.Error(err)).ServeHTTP(w, r)
	}

	headerContent, ok := r.Header["Authorization"]
	if !ok {
		w.Log(middleware.Info, blogAuthLogTag, "authorization header missing")
		onError(errMissingAuthorization, http.StatusBadRequest)
		return
	}

	authorisation := strings.Join(headerContent, " ")
	if authorisation == "" {
		w.Log(middleware.Info, blogAuthLogTag, "authorization header content is empty")
		onError(errInvalidAuthorizationContent, http.StatusBadRequest)
		return
	}

	username, password, ok := r.BasicAuth()
	if username == "" || password == "" || !ok {
		w.Log(middleware.Info, blogAuthLogTag, "given username and/or password are empty")
		onError(errInvalidAuthorizationContent, http.StatusBadRequest)
		return
	}

	h := sha512.Sum512([]byte(password))
	passwordHashed := hex.EncodeToString(h[:])

	if username != adminUsername || passwordHashed != adminPassword {
		w.Log(middleware.Info, blogAuthLogTag, "given username and/or password hash does not match administrator details")
		onError(errIncorrectCredentials, http.StatusUnauthorized)
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
			templ.Handler(components.Error(err)).ServeHTTP(w, r)
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
