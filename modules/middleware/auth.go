package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	jwt "github.com/dgrijalva/jwt-go"
)

type userKey string

const (
	// SessionName is the cookie name for sessions
	SessionName = "bmusession"
	// UserContextKey is the key for pulling user info out of request context
	UserContextKey userKey = "user"
)

// AuthConfig is configuration used for authorization operations
type AuthConfig struct {
	LegacySalt string
	JWTSecret  string
}

// validateAndDecodeAuthClaim will ensure the token provided was signed by us and decode its contents
func validateAndDecodeAuthClaim(token string, config AuthConfig) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify the algorhythm matches what we original signed
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})
	return t.Claims.(jwt.MapClaims), err
}

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

		claims, err := validateAndDecodeAuthClaim(token, AuthConfig{
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

// UserIDFromRequest retireves the user's ID from request context
func UserIDFromRequest(req *http.Request) int64 {
	return int64(req.Context().Value(UserContextKey).(jwt.MapClaims)["id"].(float64))
}
