package v1

import "net/http"

func Router() *http.ServeMux {
	r := http.NewServeMux()

	r.Handle("/github/", http.StripPrefix("/github", GithubApi()))

	return r
}
