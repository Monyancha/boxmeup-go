package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func LogHandler(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stderr, next)
}
