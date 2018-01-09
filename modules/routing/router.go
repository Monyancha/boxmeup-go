package routing

//go:generate go run ../../internal/importhooks.go ../../

import (
	"github.com/cjsaylor/boxmeup-go/hooks"
	"github.com/gorilla/mux"
)

// NewRouter gets a pre-configured router with all defined routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.Handler)
	}
	bindRouteHook(router, hooks.ImageryRouteHook{})
	return router
}

func bindRouteHook(router *mux.Router, hook hooks.RouteHook) {
	hook.Apply(router)
}
