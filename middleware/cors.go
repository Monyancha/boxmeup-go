package middleware

import (
	"net/http"

	"github.com/cjsaylor/boxmeup-go/config"
	"github.com/rs/cors"
)

func CORSHandler(next http.Handler) http.Handler {
	return corsSetup().Handler(next)
}

func CORSHandlerFunc(w http.ResponseWriter, r *http.Request) {
	corsSetup().HandlerFunc(w, r)
}

func corsSetup() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   config.Config.AllowedOrigin,
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"X-Xsrf-Token"},
		MaxAge:           600,
	})
}
