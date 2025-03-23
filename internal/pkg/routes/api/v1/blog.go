package v1

import (
	"net/http"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
)

func BlogApi() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		queries := persistence.New(utils.Database)
		posts, err := queries.GetPosts(r.Context())
		if err != nil {
			_ = components.Error().Render(r.Context(), w)
			return
		}

		_ = components.PostList(posts).Render(r.Context(), w)
	})

	return r
}
