// +build ignore

package hooks

import "github.com/gorilla/mux"

// ImageryRouteHook is the mechanism to plugin QR codes for containers
type ImageryRouteHook struct{}

// Apply of Imagery route hooks is proprietary code and generated outside this project.
func (i ImageryRouteHook) Apply(router *mux.Router) {
}
