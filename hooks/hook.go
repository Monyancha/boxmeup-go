package hooks

import (
	"github.com/gorilla/mux"
)

type RouteHook interface {
	Apply(router *mux.Router)
}
