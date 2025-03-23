package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
	w.wroteHeader = true
}

// Basic logging middleware.
//
// By default, this will log every single request received to this server in the
// following format: `received method=%s status=%d path=%s duration=%v`. For
// example, a successful request to "/" would create a log of `received
// method=GET status=0 path=/ duration=120.0µs`
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("an internal server error occurred. cause: %v\n", err)
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		log.Printf("received method=%s status=%d path=%s duration=%v\n", r.Method, wrapped.status, r.URL.EscapedPath(), time.Since(start))
	})
}
