package app

import (
	"fmt"
	"net/http"

	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type createAccountRequest struct {
	Owner    string `json:"owner" validate:"required"`
	Currency string `json:"currency" validate:"required,currency"`
}

func (server *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	arg := persistence.CreateAccountParams{
		Owner:    req.Owner,
		Balance:  0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(r.Context(), arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation, pgerrcode.ForeignKeyViolation:
				server.writeError(w, http.StatusForbidden, err)
				return
			}
		}

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

type listAccountRequest struct {
	PageID   int32 `validate:"required,min=1"`
	PageSize int32 `validate:"required,min=5,max=10"`
}

func (server *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	var req listAccountRequest
	var err error

	qs := r.URL.Query()
	err = server.readInt32(qs, "page_id", 1, &req.PageID)
	if err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	err = server.readInt32(qs, "page_size", 5, &req.PageSize)
	if err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := server.validate.Struct(req); err != nil {
		server.writeError(w, http.StatusBadRequest, fmt.Errorf("invalid page_id or page_size"))
		return
	}

	arg := persistence.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(r.Context(), arg)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	err = server.writeJSON(w, http.StatusOK, accounts, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}
