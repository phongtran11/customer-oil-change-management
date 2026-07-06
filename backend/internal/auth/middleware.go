package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

// ClaimsContextKey is the key used to store *Claims in the request context.
const ClaimsContextKey contextKey = "auth_claims"

// setClaimsInContext stores Claims in a context.
func setClaimsInContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

// ClaimsFromContext retrieves *Claims from a context.
// Returns nil if not present (route was not protected by Authenticator).
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ClaimsContextKey).(*Claims)
	return c
}

// Authenticator returns a chi-compatible middleware that validates the
// Authorization: Bearer <token> header.  On success it stores the *Claims
// value in the request context under ClaimsContextKey.
func Authenticator(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := ValidateAccessToken(parts[1], jwtSecret)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := setClaimsInContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
