package mid

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/dnmahendra/service/business/auth"
	"github.com/dnmahendra/service/foundation/web"
	"github.com/pkg/errors"
)

// ErrForbidden is returned when authenticated user does not have a
// sufficient role for an action
var ErrForbidden = web.NewRequestError(
	errors.New("you are not authorized for that action"),
	http.StatusForbidden,
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Parse the authorization header.
			// Expected header is of the format `Bearer <token>`.
			parts := strings.Split(r.Header.Get("Authorization"), " ")

			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.Key, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(log *log.Logger, roles ...string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// If the context is missing this value, request the service
			// to be shutdown gracefully
			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return errors.New("claims missing from context: HasRole called without/before Authenticate")
			}

			if !claims.Authorize(roles...) {
				log.Printf("mid: authorize: claims:  %v token: %v", claims.Roles, roles)
				return ErrForbidden
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
