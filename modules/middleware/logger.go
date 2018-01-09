package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func LogHandler(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		begin := time.Now()
		lrw := newLoggingResponseWriter(res)
		defer func() {
			fmt.Printf(
				"%v - %v [%v] \"%v %v %v\" %v %v\n",
				req.Host,
				"-",
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				req.Method,
				req.URL.EscapedPath(),
				req.Proto,
				lrw.statusCode,
				time.Since(begin),
			)
		}()
		next.ServeHTTP(lrw, req)
	}
	return http.HandlerFunc(fn)
}
