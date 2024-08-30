package api

import (
	"fmt"
	"net/http"
)

// Middleware is a function that takes an http.HandlerFunc and returns another http.HandlerFunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

type MiddlewareChain []Middleware

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
				server.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NewChain creates a new MiddlewareChain
func NewChain(middlewares ...Middleware) MiddlewareChain {
	return MiddlewareChain(middlewares)
}

// Then chains all middlewares and the final handler
func (chain *MiddlewareChain) Then(handler http.HandlerFunc) http.HandlerFunc {
	for i := range *chain {
		handler = (*chain)[len(*chain)-1-i](handler)
	}
	return handler
}
