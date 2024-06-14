package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (app *application) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (app *application) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Add("Connection", "close")
				app.serverError(w, fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
