package v1

import "net/http"

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/github/", http.StripPrefix("/github", GithubAPI()))
	mux.Handle("/blog/", http.StripPrefix("/blog", BlogAPI()))

	return mux
}
