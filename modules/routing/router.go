package routing

import (
	"fmt"
	"net/http"
	"os"
	"plugin"

	"github.com/cjsaylor/boxmeup-go/hooks"
	"github.com/cjsaylor/boxmeup-go/modules/config"
	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/items"
	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/middleware"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	"github.com/gorilla/mux"
	chain "github.com/justinas/alice"
)

var externalPlugins = []string{
	"imagery",
}

var routes = []config.Route{
	config.Route{
		Name:    "Health",
		Method:  "GET",
		Pattern: "/health",
		Handler: chain.New(middleware.LogHandler).ThenFunc(HealthHandler),
	},
}

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
	// Built in
	(users.Hook{}).Apply(router)
	(items.Hook{}).Apply(router)
	(containers.Hook{}).Apply(router)
	(locations.Hook{}).Apply(router)

	// External propriatary plugins (these assume to be in a local hooks/ folder)
	for _, name := range externalPlugins {
		plugin, err := plugin.Open(fmt.Sprintf("hooks/%s.so", name))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		routeHookSym, _ := plugin.Lookup("RouteHook")
		routeHook := routeHookSym.(*hooks.RouteHook)
		(*routeHook).Apply(router)
		fmt.Fprintf(os.Stderr, "Applied routehook: %s.so", name)
	}
	return router
}

// HealthHandler serves up a health status.
func HealthHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)
}
