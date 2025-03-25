package v1

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/types"
	"github.com/STBoyden/go-portfolio/internal/pkg/routes/site/views/components"
	gh "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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
	//nolint:revive,stylecheck // This is a struct to represent a GraphQL object, and as such needs to be named this way.
	Url         gh.URI
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
	r := http.NewServeMux()

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		panic("GITHUB_TOKEN environment variable not set!")
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	c := oauth2.NewClient(context.Background(), src)
	client := gh.NewClient(c)

	r.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		input := gh.LanguageOrder{
			Field:     gh.LanguageOrderFieldSize,
			Direction: gh.OrderDirectionDesc,
		}

		var query pinnedItemsQuery
		err := client.Query(r.Context(), &query, map[string]any{
			"languagesOrderBy": input,
		})
		if err != nil {
			fmt.Printf("an error occurred: %v\n", err)
			return
		}

		repositories := make([]types.Repository, len(query.User.PinnedItems.Nodes))
		for i, pinned := range query.User.PinnedItems.Nodes {
			entry := pinned.Entry

			repository := types.Repository{
				Owner:       string(entry.Owner.Login),
				Name:        string(entry.Name),
				Description: string(entry.Description),
				URL:         entry.Url.String(),
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

		_ = components.Repositories(repositories).Render(r.Context(), w)
	})

	return r
}
