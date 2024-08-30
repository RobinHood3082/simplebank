package api

import (
	"net/http"

	db "github.com/RobinHood3082/simplebank/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" validate:"required"`
	Currency string `json:"currency" validate:"required,oneof=EUR USD BDT GBP"`
}

func (server *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.clientError(w, http.StatusBadRequest)
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(r.Context(), arg)
	if err != nil {
		server.serverError(w, err)
		return
	}

	server.logger.Info("Account created", "account", account)
	err = server.writeJSON(w, http.StatusCreated, account, nil)
	if err != nil {
		server.serverError(w, err)
	}
}
