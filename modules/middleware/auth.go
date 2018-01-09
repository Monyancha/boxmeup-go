package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/users"
)

type userKey string

const (
	// SessionName is the cookie name for sessions
	SessionName = "bmusession"
	// UserContextKey is the key for pulling user info out of request context
	UserContextKey userKey = "user"
)

func AuthHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		sessionCookie, _ := req.Cookie(SessionName)
		var token string
		if sessionCookie != nil {
			token = sessionCookie.Value
		} else {
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(res, "Authorization required.", 401)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(res, "Authorization header must be in the form of: Bearer {token}", 401)
				return
			}
			token = parts[1]
		}

		claims, err := users.ValidateAndDecodeAuthClaim(token, users.AuthConfig{
			JWTSecret: config.Config.JWTSecret,
		})
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		if sessionCookie != nil && req.Header.Get("x-xsrf-token") != claims["xsrfToken"] {
			http.Error(res, "XSRF token mismatch!", http.StatusForbidden)
			return
		}
		var userKey userKey = "user"
		newRequest := req.WithContext(context.WithValue(req.Context(), userKey, claims))
		*req = *newRequest
		next.ServeHTTP(res, req)
		return
	}
	return http.HandlerFunc(fn)
}
