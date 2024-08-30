package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	db "github.com/RobinHood3082/simplebank/db/sqlc"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store  *db.Store
	router *Router
	logger *slog.Logger
}

// NewServer creates a new HTTP server and set up routing
func NewServer(store *db.Store, logger *slog.Logger) *Server {
	server := &Server{store: store}
	server.getRoutes()
	server.logger = logger
	return server
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(addr string) error {
	server.logger.Info(fmt.Sprintf("Starting server on %s", addr))
	return server.router.Serve(addr)
}

// The ServerError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (server *Server) ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	server.logger.Error(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The ClientError helper sends a specific status code and error message to the user.
func (server *Server) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// The NotFound helper is a convenience wrapper around clientError which sends a 404 Not Found response to the user.
func (server *Server) NotFound(w http.ResponseWriter) {
	server.ClientError(w, http.StatusNotFound)
}
