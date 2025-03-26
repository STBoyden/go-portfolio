package v1

import "net/http"

func Router() *http.ServeMux {
	r := http.NewServeMux()

	r.Handle("/github/", http.StripPrefix("/github", GithubAPI()))
	r.Handle("/blog/", http.StripPrefix("/blog", BlogAPI()))

	return r
}
