package app

import (
	"fmt"
	"net/http"
)

// LogRequest logs all incoming requests
func (server *Server) LogRequest(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.logger.Info(fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI()))
		next.ServeHTTP(w, r)
	})
}

// RecoverPanic recovers from panics and returns a 500 Internal Server Error
func (server *Server) RecoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				server.writeError(w, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
