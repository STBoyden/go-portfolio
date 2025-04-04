package v1

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	gh "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/types"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
)

type language struct {
	Color gh.String
	Name  gh.String
}

type languageEntry struct {
	Entry language `graphql:"... on Language"`
}

type languageItems struct {
	Nodes []languageEntry
}

type owner struct {
	Login gh.String
}

type repository struct {
	URL         gh.URI
	Name        gh.String
	Owner       owner
	Description gh.String
	Languages   languageItems `graphql:"languages(first: 3, orderBy: $languagesOrderBy)"`
}

type repositoryEntry struct {
	Entry repository `graphql:"... on Repository"`
}

type pinnedItems struct {
	Nodes []repositoryEntry
}

type pinnedItemsQuery struct {
	User struct {
		PinnedItems pinnedItems `graphql:"pinnedItems(first: 100, types: REPOSITORY)"`
	} `graphql:"user(login: \"STBoyden\")"`
}

func GithubAPI() *http.ServeMux {
	mux := http.NewServeMux()

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		panic("GITHUB_TOKEN environment variable not set!")
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	c := oauth2.NewClient(context.Background(), src)
	client := gh.NewClient(c)

	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		input := gh.LanguageOrder{
			Field:     gh.LanguageOrderFieldSize,
			Direction: gh.OrderDirectionDesc,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		type response struct {
			query *pinnedItemsQuery
			err   error
		}
		responsech := make(chan response)

		go func() {
			defer close(responsech)

			var query pinnedItemsQuery
			err := client.Query(ctx, &query, map[string]any{
				"languagesOrderBy": input,
			})
			if err != nil {
				responsech <- response{query: nil, err: err}
				return
			}

			responsech <- response{query: &query, err: nil}
		}()

		var query pinnedItemsQuery
	Wait:
		for {
			select {
			case <-ctx.Done():
				templ.Handler(components.Error(errors.New("timed out"))).ServeHTTP(w, r)
				return
			case resp := <-responsech:
				if resp.err != nil {
					templ.Handler(components.Error(resp.err)).ServeHTTP(w, r)
					return
				}

				query = *resp.query
				break Wait
			}
		}

		repositories := make([]types.Repository, len(query.User.PinnedItems.Nodes))
		for i, pinned := range query.User.PinnedItems.Nodes {
			entry := pinned.Entry

			repository := types.Repository{
				Owner:       string(entry.Owner.Login),
				Name:        string(entry.Name),
				Description: string(entry.Description),
				URL:         entry.URL.String(),
			}

			languages := make([]types.Language, len(entry.Languages.Nodes))
			for i, n := range entry.Languages.Nodes {
				entry := n.Entry

				language := types.Language{
					HexColour: string(entry.Color),
					Name:      string(entry.Name),
				}

				languages[i] = language
			}

			repository.Languages = languages
			repositories[i] = repository
		}

		templ.Handler(components.Repositories(repositories), templ.WithStreaming()).ServeHTTP(w, r)
	})

	return mux
}
