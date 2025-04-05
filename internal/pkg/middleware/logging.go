package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LoggerLevel interface{ loggerLevelMarker() }

type (
	_debug struct{ LoggerLevel }
	_info  struct{ LoggerLevel }
	_warn  struct{ LoggerLevel }
	_error struct{ LoggerLevel }
)

// Global logging levels for the Logging middleware.
//
//nolint:gochecknoglobals // These are logging levels and should be global.
var (
	Debug = _debug{}
	Info  = _info{}
	Warn  = _warn{}
	Error = _error{}
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
	headerPrepared    bool
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
	if statusCode == 0 {
		return
	}

	l.status = statusCode
	l.headerPrepared = true
}

// HeaderPrepared returns whether or not the header has been prepared.
func (l *LoggingMiddleware) HeaderPrepared() bool {
	return l.headerPrepared
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
	if l.wroteHeader || statusCode == 0 || l.status == statusCode {
		return
	}

	l.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
	l.wroteHeader = true
}

func (l *LoggingMiddleware) Log(level LoggerLevel, message string, attrs ...any) {
	attributes := append([]any{slog.Group("request",
		"id", l.requestID.String()[:8],
		"method", l.associatedRequest.Method,
		"status", l.status,
		"path", l.associatedRequest.URL.Path,
		"elapsed", time.Since(l.start),
	)}, attrs...)

	logger := slog.Default()

	var f func(string, ...any)
	switch level.(type) {
	case _debug:
		f = logger.Debug
	case _warn:
		f = logger.Warn
	case _error:
		f = logger.Warn
	default:
		f = logger.Info
	}

	f(message, attributes...)
}

func loggerWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wrapped *LoggingMiddleware
		defer func() {
			if err := recover(); err != nil {
				wrapped.WriteHeader(http.StatusInternalServerError)
				wrapped.Log(Error, "An internal server error occurred", "cause", err)
			}
		}()

		start := time.Now()
		wrapped = loggingWrapResponseWriter(w, r, start)

		wrapped.Log(Debug, "Request received", "received_at", start)
		next.ServeHTTP(wrapped, r)

		if wrapped.HeaderPrepared() {
			wrapped.WritePreparedHeader()
		}
	})
}
