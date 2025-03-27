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

	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/types"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
)

func blogAdmin() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /new-post/{slug}", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.AuthMiddleware](_w)

		if _, authed := w.Details(); !authed {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		reader, err := r.GetBody()
		if err != nil {
			w.PrepareHeader(http.StatusBadRequest)
			return
		}
		defer reader.Close()

		buffer, err := io.ReadAll(reader)
		if err != nil {
			w.PrepareHeader(http.StatusBadRequest)
			return
		}

		slug := r.PathValue("slug")
		if slug == "" {
			w.PrepareHeader(http.StatusBadRequest)
			return
		}

		blogContent := types.BlogContent{}
		err = json.Unmarshal(buffer, &blogContent)
		if err != nil {
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

	return r
}

func BlogAPI() *http.ServeMux {
	adminUser := utils.MustEnv("ADMIN_USER")
	adminPass := utils.MustEnv("ADMIN_PW")

	r := http.NewServeMux()

	r.HandleFunc("GET /posts", func(w http.ResponseWriter, r *http.Request) {
		queries := persistence.New(utils.Database)
		posts, err := queries.GetPosts(r.Context())
		if err != nil {
			_ = components.Error().Render(r.Context(), w)
			return
		}

		_ = components.PostList(posts).Render(r.Context(), w)
	})

	// responds with a status code relevant to the authentication status of the user.
	// 200: the user is authenticated and cookie is outside of a 30 minute expiration warning.
	// 202: the user is authenticated and cookie is within a 30 minute expiration warning.
	// 401: the user is not authenticated.
	r.HandleFunc("POST /check-authentication", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		cookie, err := r.Cookie("token")
		if err != nil {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		if cookie.Expires.Before(time.Now()) {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		} else if cookie.Expires.Before(time.Now().Add(30 * time.Minute)) {
			// if the cookie is about to expire, warn the user that they will
			// need to re-authenticate soon. use a 202 status code to indicate
			// this to the front-end.
			w.PrepareHeader(http.StatusAccepted)
		}

		token, err := uuid.Parse(cookie.Value)
		if err != nil {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		queries := persistence.New(utils.Database)
		exists, err := queries.CheckAuthExists(r.Context(), token)
		if !exists || err != nil {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}

		expired, err := queries.CheckIfAuthExpired(r.Context(), token)
		if expired || err != nil {
			w.PrepareHeader(http.StatusUnauthorized)
			return
		}
	})

	r.HandleFunc("POST /authenticate", func(_w http.ResponseWriter, r *http.Request) {
		w := utils.MustCast[middleware.LoggingMiddleware](_w)

		onError := func(statusCode int) {
			w.PrepareHeader(statusCode)
			_ = components.Error().Render(r.Context(), w)
		}

		headerContent, ok := r.Header["Authorization"]
		if !ok {
			onError(http.StatusBadRequest)
			return
		}

		authorisation := strings.Join(headerContent, " ")
		if authorisation == "" {
			onError(http.StatusBadRequest)
			return
		}

		username, password, ok := r.BasicAuth()
		if username == "" || password == "" || !ok {
			onError(http.StatusBadRequest)
			return
		}

		h := sha512.Sum512([]byte(password))
		passwordHashed := hex.EncodeToString(h[:])

		if username != adminUser || passwordHashed != adminPass {
			onError(http.StatusUnauthorized)
			return
		}

		queries := persistence.New(utils.Database)
		auth, err := queries.NewAuth(r.Context())
		if err != nil {
			panic(fmt.Sprintf("unable to create a new token: %v", err))
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   auth.ID.String(),
			Expires: auth.Expiry,
		})

		w.PrepareHeader(http.StatusOK)
	})

	r.Handle("/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/admin", blogAdmin())))

	return r
}
