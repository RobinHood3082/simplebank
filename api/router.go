package api

import (
	"fmt"
	"net/http"
)

// Router is our http router for handling different endpoints
type Router struct {
	mux *http.ServeMux
}

// NewRouter creates a new Router
func NewRouter(mux *http.ServeMux) *Router {
	return &Router{mux: mux}
}

// HandleFunc registers a new handler with a given method and path
func (r *Router) HandleFunc(method, path string, handler http.HandlerFunc) {
	pattern := fmt.Sprintf("%s %s", method, path)
	r.mux.HandleFunc(pattern, handler)
}

// Post registers a new POST request handler with the given path
func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPost, path, handler)
}

// Get registers a new GET request handler with the given path
func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodGet, path, handler)
}

// Delete registers a new DELETE request handler with the given path
func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodDelete, path, handler)
}

// Put registers a new PUT request handler with the given path
func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPut, path, handler)
}

// Patch registers a new PATCH request handler with the given path
func (r *Router) Patch(path string, handler http.HandlerFunc) {
	r.HandleFunc(http.MethodPatch, path, handler)
}

// Serve runs the http server on a specific address
func (r *Router) Serve(addr string) error {
	return http.ListenAndServe(addr, r.mux)
}
