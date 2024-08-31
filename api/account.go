package api

import (
	"fmt"
	"net/http"

	db "github.com/RobinHood3082/simplebank/db/sqlc"
	"github.com/jackc/pgx/v5"
)

type createAccountRequest struct {
	Owner    string `json:"owner" validate:"required"`
	Currency string `json:"currency" validate:"required,oneof=EUR USD BDT GBP"`
}

func (server *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(r.Context(), arg)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	server.logger.Info("Account created", "account", account)
	err = server.writeJSON(w, http.StatusCreated, account, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}

type getAccountRequest struct {
	ID int64 `validate:"required,min=1"`
}

func (server *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	var req getAccountRequest
	if _, err := fmt.Sscanf(r.PathValue("id"), "%d", &req.ID); err != nil {
		server.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid account ID"))
		return
	}

	if err := server.validate.Struct(req); err != nil {
		server.writeError(w, http.StatusBadRequest, fmt.Errorf("ID must be greater than 0"))
		return
	}

	account, err := server.store.GetAccount(r.Context(), req.ID)
	if err != nil {
		server.logger.Error(err.Error())
		if err == pgx.ErrNoRows {
			server.writeError(w, http.StatusNotFound, fmt.Errorf("account not found"))
			return
		}

		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	err = server.writeJSON(w, http.StatusOK, account, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}
