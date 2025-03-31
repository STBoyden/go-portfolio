package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/common/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/common/utils"
	"github.com/STBoyden/go-portfolio/internal/pkg/persistence"
)

// AuthMiddleware is an extension over [LoggingMiddleware] (and by extension
// [http.ResponseWriter]) to handle proper authorisation of a child
// [http.Handler].
//
// For AuthMiddleware to work, a parent [http.Handler] must have been wrapped by
// the Logger wrapper method as it is required for this middleware.
type AuthMiddleware struct {
	*LoggingMiddleware

	authed bool
	token  uuid.UUID
}

// Details returns a pair, a nullable [uuid.UUID] and bool depending on whether
// or not the user is authorised with the [uuid.UUID] representing the
// authorised token previously checked.
func (a AuthMiddleware) Details() (*uuid.UUID, bool) {
	if a.authed {
		return &a.token, true
	}

	return nil, false
}

func (a AuthMiddleware) Authed() bool {
	return a.authed
}

// Wrapper over [LoggingMiddleware.Log] to standardise the prefix.
func (a *AuthMiddleware) Log(level level, format string, v ...any) {
	a.LoggingMiddleware.Log(level, "auth", format, v...)
}

func (a *AuthMiddleware) Authorise(r *http.Request) bool {
	ctx := r.Context()

	cookie, err := r.Cookie(consts.TokenCookieName)
	if err != nil {
		a.Log(Error, "authorisation failed: missing cookie")

		return false
	}

	token, err := uuid.Parse(cookie.Value)
	if err != nil {
		a.Log(Error, "authorisation failed: incorrect token format")

		return false
	}

	queries := persistence.New(utils.Database)

	exists, err := queries.CheckAuthExists(ctx, token)
	if !exists || err != nil {
		a.Log(Error, "authorisation failed: token %v does not exist in database", token)

		return false
	}

	expired, err := queries.CheckIfAuthExpired(ctx, token)
	if expired || err != nil {
		a.Log(Error, "authorisation failed: token %v's authorisation has expired", token)

		return false
	}

	a.token = token
	a.authed = true

	return true
}

func authWrapResponseWriter(l *LoggingMiddleware) *AuthMiddleware {
	return &AuthMiddleware{LoggingMiddleware: l}
}

func authorisationWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(__w http.ResponseWriter, r *http.Request) {
		_w := utils.MustCast[LoggingMiddleware](__w)
		w := authWrapResponseWriter(_w)

		w.Log(Info, "checking authorisation for path=%s", r.URL.EscapedPath())

		authed := w.Authorise(r)
		if !authed {
			w.PrepareHeader(http.StatusUnauthorized)
		}

		next.ServeHTTP(w, r)
	})
}
