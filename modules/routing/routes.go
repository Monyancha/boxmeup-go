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
		"Login",
		"POST",
		"/api/user/login",
		chain.New(middleware.LogHandler, middleware.JsonResponseHandler).ThenFunc(LoginHandler),
	},
	Route{
		"Logout",
		"GET",
		"/api/user/logout",
		chain.New(middleware.LogHandler).ThenFunc(LogoutHandler),
	},
	Route{
		"Register",
		"POST",
		"/api/user/register",
		chain.New(middleware.LogHandler, middleware.JsonResponseHandler).ThenFunc(RegisterHandler),
	},
	Route{
		"User",
		"GET",
		"/api/user/current",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(UserHandler),
	},
	Route{
		"CreateContainer",
		"POST",
		"/api/container",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(CreateContainerHandler),
	},
	Route{
		"UpdateContainer",
		"PUT",
		"/api/container/{id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(UpdateContainerHandler),
	},
	Route{
		"DeleteContainer",
		"DELETE",
		"/api/container/{id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(DeleteContainerHandler),
	},
	Route{
		"Container",
		"GET",
		"/api/container/{id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(ContainerHandler),
	},
	Route{
		"Containers",
		"GET",
		"/api/container",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(ContainersHandler),
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
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"ModifyContainerItem",
		"PUT",
		"/api/container/{id}/item/{item_id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(SaveContainerItemHandler),
	},
	Route{
		"DeleteItems",
		"DELETE",
		"/api/container/{id}/item/{item_id}",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(DeleteContainerItemHandler),
	},
	Route{
		"Items",
		"GET",
		"/api/container/{id}/item",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(ContainerItemsHandler),
	},
	Route{
		"Items",
		"GET",
		"/api/item/search",
		chain.New(middleware.LogHandler, middleware.AuthHandler, middleware.JsonResponseHandler).ThenFunc(SearchItemHandler),
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
