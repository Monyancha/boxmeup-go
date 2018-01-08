package routing

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	chain "github.com/justinas/alice"
)

// SessionName is the cookie name for sessions
const SessionName = "bmusession"

// Route defines a route
type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

// Routes is a collection of routes
type Routes []Route

type userKey string

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func logHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		begin := time.Now()
		lrw := newLoggingResponseWriter(res)
		defer func() {
			fmt.Printf(
				"%v - %v [%v] \"%v %v %v\" %v %v\n",
				req.Host,
				"-",
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				req.Method,
				req.URL.EscapedPath(),
				req.Proto,
				lrw.statusCode,
				time.Since(begin),
			)
		}()
		next.ServeHTTP(lrw, req)
	}
	return http.HandlerFunc(fn)
}

func jsonResponseHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}

func authHandler(next http.Handler) http.Handler {
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

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		chain.New(logHandler).ThenFunc(IndexHandler),
	},
	Route{
		"Health",
		"GET",
		"/health",
		chain.New(logHandler).ThenFunc(HealthHandler),
	},
	Route{
		"Login",
		"POST",
		"/api/user/login",
		chain.New(logHandler, jsonResponseHandler).ThenFunc(LoginHandler),
	},
	Route{
		"Logout",
		"GET",
		"/api/user/logout",
		chain.New(logHandler).ThenFunc(LogoutHandler),
	},
	Route{
		"Register",
		"POST",
		"/api/user/register",
		chain.New(logHandler, jsonResponseHandler).ThenFunc(RegisterHandler),
	},
	Route{
		"User",
		"GET",
		"/api/user/current",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(UserHandler),
	},
	Route{
		"CreateContainer",
		"POST",
		"/api/container",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(CreateContainerHandler),
	},
	Route{
		"UpdateContainer",
		"PUT",
		"/api/container/{id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(UpdateContainerHandler),
	},
	Route{
		"DeleteContainer",
		"DELETE",
		"/api/container/{id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(DeleteContainerHandler),
	},
	Route{
		"Container",
		"GET",
		"/api/container/{id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(ContainerHandler),
	},
	Route{
		"Containers",
		"GET",
		"/api/container",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(ContainersHandler),
	},
	Route{
		"ContainerQR",
		"GET",
		"/api/container/{id}/qrcode",
		chain.New(logHandler, authHandler).ThenFunc(ContainerQR),
	},
	Route{
		"CreateContainerItem",
		"POST",
		"/api/container/{id}/item",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"ModifyContainerItem",
		"PUT",
		"/api/container/{id}/item/{item_id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"DeleteItems",
		"DELETE",
		"/api/container/{id}/item/{item_id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(DeleteContainerItemHandler),
	},
	Route{
		"Items",
		"GET",
		"/api/container/{id}/item",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(ContainerItemsHandler),
	},
	Route{
		"Items",
		"GET",
		"/api/item/search",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(SearchItemHandler),
	},
	Route{
		"CreateLocation",
		"POST",
		"/api/location",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(CreateLocationHandler),
	},
	Route{
		"UpdateLocation",
		"PUT",
		"/api/location/{id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(UpdateLocationHandler),
	},
	Route{
		"DeleteLocation",
		"DELETE",
		"/api/location/{id}",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(DeleteLocationHandler),
	},
	Route{
		"Locations",
		"GET",
		"/api/location",
		chain.New(logHandler, authHandler, jsonResponseHandler).ThenFunc(LocationsHandler),
	},
}
