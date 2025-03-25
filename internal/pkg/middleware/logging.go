package middleware

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

// LoggingMiddleware is an extension over [http.ResponseWriter] to standardise
// logging throughout the application. If the [http.Handler] has been wrapped in
// the Logger wrapper function, then all children will be able to cast their
// [http.ResponseWriter] types to LoggingMiddleware and use
// [LoggingMiddleware.Log] directly. Additionally, middleware's that require
// LoggingMiddleware can implement a Log method to modify the default behaviour.
// For example, see [AuthMiddleware.Log].
type LoggingMiddleware struct {
	http.ResponseWriter

	start             time.Time
	associatedRequest *http.Request
	requestID         uuid.UUID
	status            int
	wroteHeader       bool
}

var _ middleware = (*LoggingMiddleware)(nil)

func loggingWrapResponseWriter(w http.ResponseWriter, r *http.Request, s time.Time) *LoggingMiddleware {
	return &LoggingMiddleware{ResponseWriter: w, associatedRequest: r, requestID: uuid.New(), start: s}
}

// Status returns the current status of the request.
func (l *LoggingMiddleware) Status() int {
	return l.status
}

// PrepareHeader sets the status code of the request without calling
// [LoggingMiddleware.WriteHeader].
func (l *LoggingMiddleware) PrepareHeader(statusCode int) {
	l.status = statusCode
}

// WritePreparedHeader writes the previously prepared w.status to the status
// header for the request. Subsequent calls to
// [LoggingMiddleware.WritePreparedHeader] or [LoggingMiddleware.WriteHeader]
// are superflous and will be ignored.
func (l *LoggingMiddleware) WritePreparedHeader() {
	l.WriteHeader(l.status)
}

// WriteHeader writes the given statusCode to the status header for the request.
// Subsequent calls to [LoggingMiddleware.WritePreparedHeader] or
// [LoggingMiddleware.WriteHeader] are superflous and will be ignored.
func (l *LoggingMiddleware) WriteHeader(statusCode int) {
	if l.wroteHeader {
		return
	}

	l.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
	l.wroteHeader = true
}

func (l *LoggingMiddleware) Log(prefix, format string, v ...any) {
	f := "request_id=%v method=%s status=%s path=%s elapsed=%v"
	if format != "" {
		f = fmt.Sprintf("%s\n\t[%s] msg=%s", f, prefix, format)
	}

	status := fmt.Sprintf("%d", l.status)
	if l.status == 0 {
		status = "pending"
	}

	args := []any{
		l.requestID.String()[:8],
		l.associatedRequest.Method,
		status,
		l.associatedRequest.URL.EscapedPath(),
		time.Since(l.start),
	}
	args = slices.Concat(args, v)

	log.Printf(f, args...)
}

func loggerWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wrapped *LoggingMiddleware
		defer func() {
			if err := recover(); err != nil {
				wrapped.Log("http", "before error: %v", wrapped.wroteHeader)
				wrapped.WriteHeader(http.StatusInternalServerError)
				wrapped.Log("http", "an internal server error occurred. cause: %v", err)
			}
		}()

		start := time.Now()
		wrapped = loggingWrapResponseWriter(w, r, start)
		next.ServeHTTP(wrapped, r)
		wrapped.WritePreparedHeader()

		wrapped.Log("http", "finished handling request")
	})
}
