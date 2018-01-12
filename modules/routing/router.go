package routing

import (
	"fmt"
	"os"
	"plugin"

	"github.com/cjsaylor/boxmeup-go/hooks"
	"github.com/cjsaylor/boxmeup-go/modules/items"
	"github.com/gorilla/mux"
)

var externalPlugins = []string{
	"imagery",
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
	(items.Hook{}).Apply(router)

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
