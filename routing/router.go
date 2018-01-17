package routing

import (
	"fmt"
	"net/http"
	"os"
	"plugin"

	"github.com/cjsaylor/boxmeup-go/hooks"
	"github.com/cjsaylor/boxmeup-go/middleware"
	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/items"
	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/users"
	"github.com/gorilla/mux"
)

var externalPlugins = []string{
	"imagery",
}

// NewRouter gets a pre-configured router with all defined routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Add a CORS pre-flight handler
	router.Methods("OPTIONS").HandlerFunc(middleware.CORSHandlerFunc)

	// Add a health endpoint
	// @todo check a thread safe shutdown flag as well as look at database health
	router.Methods("GET").Path("/health").HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNoContent)
	})

	// Global middleware
	router.Use(middleware.CORSHandler)
	router.Use(middleware.LogHandler)

	// Built in
	(users.Hook{}).Apply(router)
	(items.Hook{}).Apply(router)
	(containers.Hook{}).Apply(router)
	(locations.Hook{}).Apply(router)

	// External propriatary plugins (these assume to be in a local hooks/ folder)
	loadExternalPlugins(router)

	return router
}

func loadExternalPlugins(router *mux.Router) {
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
}
