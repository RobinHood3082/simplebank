package api

import "net/http"

func (server *Server) getRoutes() {
	router := NewRouter(http.NewServeMux())

	standard := NewChain(server.LogRequest, server.RecoverPanic)

	router.Get("/health", standard.Then(server.healthCheck))
	router.Post("/accounts", standard.Then(server.createAccount))
	router.Get("/accounts/{id}", standard.Then(server.getAccount))

	server.router = router
}

func (server *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("HEALTH OK"))
}
