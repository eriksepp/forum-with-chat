package log

import (
	"net/http"

	"forum/application"
	"forum/logger"
)

func Log(app *application.Application, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.InfoLog.Printf("Request URL: '%v' Method: '%v' Remote address: '%v'", r.URL, r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func TestOne(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware Test One Before")
		next.ServeHTTP(w, r)
		logger.Info("Middleware Test One After")
	})
}

func TestTwo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Middleware Test Two Before")
		next.ServeHTTP(w, r)
		logger.Info("Middleware Test Two After")
	})
}
