package middleware

import "net/http"

// Middleware is a function that takes an http.HandlerFunc and returns another http.HandlerFunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

// MiddlewareChain is a list of middlewares
type MiddlewareChain []Middleware

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
