package app

import (
	"fmt"
	"net/http"

	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/internal/token"
	"github.com/jackc/pgx/v5"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" validate:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" validate:"required,min=1"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	Currency      string `json:"currency" validate:"required,currency"`
}

func (server *Server) createTransfer(w http.ResponseWriter, r *http.Request) {
	var req createTransferRequest
	if err := server.bindData(w, r, &req); err != nil {
		server.writeError(w, http.StatusBadRequest, err)
		return
	}

	fromAccount, valid := server.validAccount(w, r, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := r.Context().Value(AuthorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		server.writeError(w, http.StatusUnauthorized, fmt.Errorf("account doesn't belong to the authenticated user"))
		return
	}

	_, valid = server.validAccount(w, r, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := persistence.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	Transfer, err := server.store.TransferTx(r.Context(), arg)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
		return
	}

	server.logger.Info("Transfer created", "Transfer", Transfer)
	err = server.writeJSON(w, http.StatusOK, Transfer, nil)
	if err != nil {
		server.writeError(w, http.StatusInternalServerError, err)
	}
}

func (server *Server) validAccount(w http.ResponseWriter, r *http.Request, accountID int64, currency string) (persistence.Account, bool) {
	account, err := server.store.GetAccount(r.Context(), accountID)
	if err != nil {
		server.logger.Error(err.Error())
		if err == pgx.ErrNoRows {
			server.writeError(w, http.StatusNotFound, fmt.Errorf("account with ID %d not found", accountID))
			return account, false
		}

		server.writeError(w, http.StatusInternalServerError, err)
		return account, false
	}

	if account.Currency != currency {
		server.writeError(w, http.StatusBadRequest, fmt.Errorf("account currency mismatch: expected %s, got %s", account.Currency, currency))
		return account, false
	}

	return account, true
}
