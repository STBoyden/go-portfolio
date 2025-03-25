package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/types"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/middleware"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
	"github.com/google/uuid"
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

		w.Header().Set("content-type", "application/json")
		err = json.NewEncoder(w).Encode(response{PostID: post.ID})
		if err != nil {
			panic(fmt.Sprintf("unable to create response object: %v", err))
		}
	})

	return r
}

func BlogApi() *http.ServeMux {
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

	r.Handle("/admin/", middleware.Handlers.Authorisation(http.StripPrefix("/admin", blogAdmin())))

	return r
}
