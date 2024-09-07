package app

import (
	"net/http"

	"github.com/RobinHood3082/simplebank/pkg/middleware"
	"github.com/RobinHood3082/simplebank/pkg/router"
)

func (server *Server) getRoutes() {
	router := router.NewRouter(http.NewServeMux())

	standard := middleware.NewChain(server.LogRequest, server.RecoverPanic)

	router.Get("/health", standard.Then(server.healthCheck))
	router.Post("/accounts", standard.Then(server.createAccount))
	router.Get("/accounts/{id}", standard.Then(server.getAccount))
	router.Get("/accounts", standard.Then(server.listAccounts))

	server.router = router
}

func (server *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("HEALTH OK"))
}
