package middleware

import "net/http"

func JsonResponseHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}
