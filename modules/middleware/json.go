package middleware

import "net/http"

type JsonErrorResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func JsonResponseHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=UTF-8")
		next.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}
