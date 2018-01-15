package routing

import (
	"fmt"
	"net/http"
)

// IndexHandler serves the static page
func IndexHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Welcome!")
}

// HealthHandler serves up a health status.
func HealthHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)
}
