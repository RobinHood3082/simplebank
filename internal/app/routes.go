package app

import (
	"net/http"

	"github.com/RobinHood3082/simplebank/pkg/middleware"
	"github.com/RobinHood3082/simplebank/pkg/router"
)

func (server *Server) getRoutes() {
	router := router.NewRouter(http.NewServeMux())

	standardChain := middleware.NewChain(server.LogRequest, server.RecoverPanic)
	authenticatedChain := append(standardChain, server.Authenticate)

	router.Get("/health", standardChain.Then(server.healthCheck))

	router.Post("/users", standardChain.Then(server.createUser))
	router.Post("/users/login", standardChain.Then(server.loginUser))
	router.Post("/tokens/renew_access", standardChain.Then(server.renewAccessToken))

	router.Post("/accounts", authenticatedChain.Then(server.createAccount))
	router.Get("/accounts/{id}", authenticatedChain.Then(server.getAccount))
	router.Get("/accounts", authenticatedChain.Then(server.listAccounts))

	router.Post("/transfers", authenticatedChain.Then(server.createTransfer))

	server.router = router
}

func (server *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("HEALTH OK"))
}
