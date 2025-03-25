// middleware package contains useful middleware to wrap over routes to provide
// useful functionality.
package middleware

import "net/http"

type (
	middleware        interface{ http.ResponseWriter }
	MiddlewareWrapper func(next http.Handler) http.Handler
)

type middlewares struct {
	// Basic logging middleware.
	//
	// By default, this will log every single request received to this server in the
	// following format:
	//
	//	request_id=%s method=%s status=%s path=%s elapsed=%v
	//
	// For example, a successful request to "/" would create a log of:
	//
	//	request_id=abcd1234 method=GET status=200 path=/ elapsed=120.0µs
	//
	// # Child middlewares
	//
	// Child middlewares can be created by embedding *[LoggingMiddleware] in the
	// implementor struct fields. Assuming that the [LoggingMiddleware] middleware
	// is correctly wrapping the parent http.Handler, child middlewares can cast the
	// given http.ResponseWriter from the closure parameter in http.HandlerFunc to a
	// *[LoggingMiddleware]. For example, through [utils.MustCast]:
	//
	//	wrapped := utils.MustCast[*middleware.LoggingMiddleware](w /* http.ResponseWriter */)
	//
	// Or:
	//
	//	wrapped, ok := w.(*middleware.LoggingMiddleware)
	//	if !ok { /* handle potential that it's not castable */ }
	//	// use otherwise
	//
	// Child middlewares can reimplement the Log method to have a consistent prefix
	// for the messsage portion of the log which is output on a separate line.See
	// [middleware.AuthMiddleware] for an example of implementation.
	Logger MiddlewareWrapper

	// Authorisation middleware handles authentication of requests for paths
	// handle by the given next [http.Handler].
	//
	// Authorisation requires that the Logger middleware wrapper over or is a
	// parent [http.Handler].
	Authorisation MiddlewareWrapper
}

var Handlers = &middlewares{
	Logger:        loggerWrapper,
	Authorisation: authorisationWrapper,
}
