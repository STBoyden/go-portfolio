package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/STBoyden/go-portfolio/internal/pkg/consts"
	"github.com/STBoyden/go-portfolio/internal/pkg/utils"
)

// AuthMiddleware is an extension over [LoggingMiddleware] (and by extension
// [http.ResponseWriter]) to handle proper authorisation of a child
// [http.Handler].
//
// For AuthMiddleware to work, a parent [http.Handler] must have been wrapped by
// the Logger wrapper method as it is required for this middleware.
//
//nolint:recvcheck // The Authed, Details and Log methods do not need to have pointer receivers as they should not have the ability to modify the struct.
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
func (a AuthMiddleware) Log(level LoggerLevel, message string, attrs ...any) {
	group := slog.Group("auth", attrs...)
	a.LoggingMiddleware.Log(level, message, group)
}

func (a *AuthMiddleware) Authorise(r *http.Request) bool {
	ctx := r.Context()

	cookie, err := r.Cookie(consts.TokenCookieName)
	if err != nil {
		a.Log(Error, "Authorisation failed: missing cookie")

		return false
	}

	token, err := uuid.Parse(cookie.Value)
	if err != nil {
		a.Log(Error, "Authorisation failed: incorrect token format")

		return false
	}

	queries, commit, rollback, err := utils.Database.StartReadTx(r.Context())
	if err != nil {
		panic("unable to start transaction on database: " + err.Error())
	}
	defer rollback(r.Context())

	exists, err := queries.CheckAuthExists(ctx, token)
	if !exists || err != nil {
		a.Log(Error, "Authorisation failed: token does not exist in database", "token", token)

		return false
	}

	expired, err := queries.CheckIfAuthExpired(ctx, token)
	if expired || err != nil {
		a.Log(Error, "Authorisation failed: token's authorisation has expired", "token", token)

		return false
	}

	a.token = token
	a.authed = true

	_ = commit(r.Context())

	return true
}

func authWrapResponseWriter(l *LoggingMiddleware) *AuthMiddleware {
	return &AuthMiddleware{LoggingMiddleware: l}
}

func authorisationWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(__w http.ResponseWriter, r *http.Request) {
		_w := utils.MustCast[LoggingMiddleware](__w)
		w := authWrapResponseWriter(_w)

		authed := w.Authorise(r)
		if !authed {
			w.PrepareHeader(http.StatusUnauthorized)
		}

		w.Log(Info, "Checking authorisation for path", "authorised", authed)

		next.ServeHTTP(w, r)
	})
}
