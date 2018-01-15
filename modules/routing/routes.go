package routing

import (
	"net/http"

	"github.com/cjsaylor/boxmeup-go/modules/middleware"
	chain "github.com/justinas/alice"
)

// Route defines a route
type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.Handler
}

// Routes is a collection of routes
type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		chain.New(middleware.LogHandler).ThenFunc(IndexHandler),
	},
	Route{
		"Health",
		"GET",
		"/health",
		chain.New(middleware.LogHandler).ThenFunc(HealthHandler),
	},
	Route{
		"CreateLocation",
		"POST",
		"/api/location",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(CreateLocationHandler),
	},
	Route{
		"UpdateLocation",
		"PUT",
		"/api/location/{id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(UpdateLocationHandler),
	},
	Route{
		"DeleteLocation",
		"DELETE",
		"/api/location/{id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(DeleteLocationHandler),
	},
	Route{
		"Locations",
		"GET",
		"/api/location",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(LocationsHandler),
	},
}
